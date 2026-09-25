package game

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// ability_ref_triggered_test.go — ADR 0041 phase 3 tier 4, slice 4-2
// (P9, #1497): a triggered row that DECLARES its Effect is built by the
// engine and named on its stack item, so a table with the trigger
// waiting is a restore point. Restore takes the row back through the
// same Q2 name check and the same Q3 manual-item policy the activated
// slot uses. Like ability_ref_test.go, every test swaps CatalogLookup
// between the capture and the restore, which is what a deploy does.

const trigRefOracle = "fixture-trigger-ref"

const (
	trigGainKey  = "Trig Probe — you gain 1 life"
	trigDrainKey = "Trig Probe — each opponent loses 1 life"
)

func trigYourUpkeep(ev Event, source *Card, _ Characteristic, _ *Game) bool {
	return ev.Actor == source.Controller
}

// trigGainRow watches draws, so the upkeep the tests fire never
// triggers it: it is there to be the row a reorder moves.
func trigGainRow() TriggeredAbility {
	return TriggeredAbility{
		Watches:   []EventKind{EventDrawCard},
		AppliesTo: trigYourUpkeep,
		Key:       trigGainKey,
		Effect: func(g *Game, item *StackItem) error {
			return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, 1)
		},
	}
}

func trigDrainRow(key string) TriggeredAbility {
	return TriggeredAbility{
		Watches:   []EventKind{EventBeginUpkeep},
		AppliesTo: trigYourUpkeep,
		Key:       key,
		Effect: func(g *Game, item *StackItem) error {
			for _, p := range g.Seats {
				if p.ID != item.Controller {
					if err := g.ChangePlayerLifeForEffect(item.SourceCardID, p.ID, -1); err != nil {
						return err
					}
				}
			}
			return nil
		},
	}
}

// withTrigCatalog files a one-card catalog whose triggered rows are
// `rows`, stamped with their identity exactly as the registry does.
func withTrigCatalog(t *testing.T, rows ...TriggeredAbility) {
	t.Helper()
	d := &CardDef{Triggered: rows}
	IdentifyCatalogRows(trigRefOracle, d)
	withCatalog(t, trigRefOracle, d)
}

// fireTrigProbe puts the probe on seat 0's battlefield and fires seat
// 0's upkeep, so the drain row triggers. It returns the probe and the
// queued item.
func fireTrigProbe(t *testing.T, g *Game) (uuid.UUID, *StackItem) {
	t.Helper()
	me := g.Seats[0].ID
	probe := pushTypedTestCard(g, Card{
		Name: "Trig Probe", TypeLine: "Enchantment", OracleID: trigRefOracle,
		Owner: me, Controller: me,
	})
	var item *StackItem
	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventBeginUpkeep, Actor: me})
		for _, it := range g.PendingTriggers {
			if it.SourceCardID == probe {
				item = it
			}
		}
	})
	if item == nil {
		t.Fatal("the upkeep queued no trigger for the probe")
	}
	return probe, item
}

// restoredPendingTrigger is the one queued trigger of a restored game.
func restoredPendingTrigger(t *testing.T, g *Game) *StackItem {
	t.Helper()
	var out *StackItem
	g.ReadSnapshot(func() {
		if len(g.PendingTriggers) == 1 {
			out = g.PendingTriggers[0]
		}
	})
	if out == nil {
		t.Fatal("the restored game has no queued trigger")
	}
	return out
}

// resolveQueuedTrigger puts the queued trigger on the stack and
// resolves it.
func resolveQueuedTrigger(g *Game) {
	g.WithWriteLock(func() {
		g.drainPendingTriggersAPNAPLocked()
		g.resolveTopAbilityLocked()
	})
}

func TestDeclaredTriggerIsStampedAndRestores(t *testing.T) {
	withTrigCatalog(t, trigGainRow(), trigDrainRow(trigDrainKey))
	g := newActiveGame(t)
	_, item := fireTrigProbe(t, g)

	want := AbilityRef{Key: trigRefOracle, Slot: AbilitySlotTriggered, Ref: "own:1", Name: trigDrainKey}
	if item.Body != CatalogTriggeredBodyKey || item.Params.Ability == nil || *item.Params.Ability != want {
		t.Fatalf("item stamp = body %q ref %+v, want %q %+v", item.Body, item.Params.Ability, CatalogTriggeredBodyKey, want)
	}
	if item.Label != trigDrainKey {
		t.Errorf("the engine built the item as %q, want the row's Key %q", item.Label, trigDrainKey)
	}
	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("a declared trigger waiting to go on the stack blocks the restore point: %+v", snap.Continuations)
	}
	restored, err := throughJSON(t, snap).RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	back := restoredPendingTrigger(t, restored)
	if back.Effect == nil || back.Body != CatalogTriggeredBodyKey {
		t.Fatalf("restore did not re-derive the row: effect %v body %q", back.Effect != nil, back.Body)
	}
	opp := restored.Seats[1]
	before := opp.Life
	resolveQueuedTrigger(restored)
	if opp.Life != before-1 {
		t.Errorf("the restored trigger resolved to life %d, want %d", opp.Life, before-1)
	}
}

// Q2: a deploy reordered the card's triggers; the one row that still
// carries the declared Key is the ability.
func TestTriggerRefFollowsItsKeyAfterAReorder(t *testing.T) {
	withTrigCatalog(t, trigGainRow(), trigDrainRow(trigDrainKey))
	g := newActiveGame(t)
	probe, _ := fireTrigProbe(t, g)
	snap := throughJSON(t, g.CaptureSnapshot())

	withTrigCatalog(t, trigDrainRow(trigDrainKey), trigGainRow())
	if lost := snap.LostStackAbilities(); len(lost) != 0 {
		t.Fatalf("a reordered row is reported lost: %+v", lost)
	}
	restored, err := snap.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	back := restoredPendingTrigger(t, restored)
	if back.Params.Ability == nil || back.Params.Ability.Ref != "own:0" {
		t.Fatalf("ref after the reorder = %+v, want own:0", back.Params.Ability)
	}
	opp, me := restored.Seats[1], restored.Seats[0]
	oppBefore, meBefore := opp.Life, me.Life
	resolveQueuedTrigger(restored)
	if opp.Life != oppBefore-1 || me.Life != meBefore {
		t.Errorf("after resolving: opponent %d -> %d, me %d -> %d; want the drain, not the gain",
			oppBefore, opp.Life, meBefore, me.Life)
	}
	restored.ReadSnapshot(func() {
		if c := restored.findCardByIDLocked(probe); c == nil || c.AbilitiesLostOnRestore {
			t.Error("a card whose trigger was found is flagged lost")
		}
	})
}

// Q3: no row carries the Key any more. The game restores; the trigger
// stays queued as a manual item; the card is flagged; the boot report
// names it.
func TestTriggerThatNoRowAnswersRestoresAsAManualItem(t *testing.T) {
	withTrigCatalog(t, trigGainRow(), trigDrainRow(trigDrainKey))
	g := newActiveGame(t)
	probe, item := fireTrigProbe(t, g)
	itemID := item.ID
	snap := throughJSON(t, g.CaptureSnapshot())

	withTrigCatalog(t, trigGainRow(), trigDrainRow(trigDrainKey+" (reworded)"))
	lost := snap.LostStackAbilities()
	if len(lost) != 1 || lost[0].ItemID != itemID || lost[0].CardID != probe || lost[0].Card != "Trig Probe" ||
		lost[0].Ref.Slot != AbilitySlotTriggered || lost[0].Ref.Name != trigDrainKey {
		t.Fatalf("LostStackAbilities = %+v, want the probe's trigger", lost)
	}
	restored, err := snap.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict refused the table over one trigger: %v", err)
	}
	back := restoredPendingTrigger(t, restored)
	if back.ID != itemID || back.Effect != nil || back.Body != "" || back.Params.Ability != nil ||
		back.targetSpec != nil || back.modeSpec != nil {
		t.Errorf("the lost trigger came back as %+v, want a manual item with nothing to run", back)
	}
	if back.Label != trigDrainKey {
		t.Errorf("the manual item lost its label: %q", back.Label)
	}
	restored.ReadSnapshot(func() {
		if c := restored.findCardByIDLocked(probe); c == nil || !c.AbilitiesLostOnRestore {
			t.Error("the source card is not flagged AbilitiesLostOnRestore")
		}
	})
	opp := restored.Seats[1]
	before := opp.Life
	resolveQueuedTrigger(restored)
	if opp.Life != before {
		t.Errorf("the manual item did something: life %d -> %d", before, opp.Life)
	}
	again := restored.CaptureSnapshot()
	if !again.Restorable() || len(again.LostStackAbilities()) != 0 {
		t.Errorf("the recapture: restorable %v, lost %+v", again.Restorable(), again.LostStackAbilities())
	}
}

// A fill-in Build supplies what the engine cannot know — here a Params
// value read at trigger time — and leaves the Effect to the row. The
// value rides the item across a restore.
func TestFillInBuildParamsSurviveARestore(t *testing.T) {
	row := TriggeredAbility{
		Watches:   []EventKind{EventBeginUpkeep},
		AppliesTo: trigYourUpkeep,
		Key:       "Trig Probe — lose the amount",
		Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
			item := NewTriggeredItem(source, "Trig Probe — lose 3", nil)
			item.Params.Amount = 3
			return item
		},
		Effect: func(g *Game, item *StackItem) error {
			return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -item.Params.Amount)
		},
	}
	withTrigCatalog(t, row)
	g := newActiveGame(t)
	_, item := fireTrigProbe(t, g)
	if item.Body != CatalogTriggeredBodyKey || item.Params.Amount != 3 || item.Label != "Trig Probe — lose 3" {
		t.Fatalf("item = body %q amount %d label %q", item.Body, item.Params.Amount, item.Label)
	}
	restored, err := throughJSON(t, g.CaptureSnapshot()).RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	me := restored.Seats[0]
	before := me.Life
	resolveQueuedTrigger(restored)
	if me.Life != before-3 {
		t.Errorf("life %d -> %d, want -3 from the carried Params", before, me.Life)
	}
}

// The legacy shape — a Build that computes its effect, no Effect beside
// it — is not stamped, and blocks the restore point as it always did.
func TestLegacyTriggerBuildIsNotStamped(t *testing.T) {
	row := trigDrainRow(trigDrainKey)
	effect := row.Effect
	row.Effect = nil
	row.Build = func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
		return NewTriggeredItem(source, trigDrainKey, effect)
	}
	withTrigCatalog(t, row)
	g := newActiveGame(t)
	_, item := fireTrigProbe(t, g)
	if item.Body != "" || item.Params.Ability != nil {
		t.Errorf("a legacy Build was stamped: %q %+v", item.Body, item.Params.Ability)
	}
	if snap := g.CaptureSnapshot(); snap.Continuations.StackEffects != 1 {
		t.Errorf("census = %+v, want the unkeyed item counted once in StackEffects", snap.Continuations)
	}
}

// A row whose TargetsFrom reads the board is not stamped either: its
// clause could not be rebuilt exactly on the restored board.
func TestTargetsFromReadsBoardIsNotStamped(t *testing.T) {
	row := trigDrainRow(trigDrainKey)
	row.TargetsFromReadsBoard = true
	withTrigCatalog(t, row)
	g := newActiveGame(t)
	_, item := fireTrigProbe(t, g)
	if item.Body != "" || item.Params.Ability != nil {
		t.Errorf("a board-reading row was stamped: %q %+v", item.Body, item.Params.Ability)
	}
}

// A row the catalog did not file — an engine trigger, a stub — builds
// from its Effect but has no name to stamp.
func TestUnfiledDeclaredTriggerIsNotStamped(t *testing.T) {
	d := &CardDef{Triggered: []TriggeredAbility{trigDrainRow(trigDrainKey)}}
	withCatalog(t, trigRefOracle, d)
	g := newActiveGame(t)
	_, item := fireTrigProbe(t, g)
	if item.Body != "" || item.Params.Ability != nil || item.Effect == nil {
		t.Errorf("unfiled row: body %q ref %+v effect %v", item.Body, item.Params.Ability, item.Effect != nil)
	}
}

// A Build beside a declared Effect must leave item.Effect alone: in a
// test binary the mistake panics at the first trigger.
func TestFillInBuildThatSetsTheEffectIsAFault(t *testing.T) {
	row := trigDrainRow(trigDrainKey)
	effect := row.Effect
	row.Build = func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
		return NewTriggeredItem(source, trigDrainKey, effect)
	}
	withTrigCatalog(t, row)
	g := newActiveGame(t)
	defer func() {
		r := recover()
		if r == nil || !strings.Contains(r.(string), "fill-in Build must leave it nil") {
			t.Errorf("recover() = %v, want the effectKeyFault", r)
		}
	}()
	fireTrigProbe(t, g)
}

// P10: a binary from 4-1 does not register catalog/triggered and
// refuses a file that names it.
func TestABinaryWithoutTheTriggeredBodyRefusesAStampedTrigger(t *testing.T) {
	withTrigCatalog(t, trigGainRow(), trigDrainRow(trigDrainKey))
	g := newActiveGame(t)
	fireTrigProbe(t, g)
	snap := throughJSON(t, g.CaptureSnapshot())
	if _, err := snap.RestoreStrict(); err != nil {
		t.Fatalf("control: this binary refuses its own file: %v", err)
	}
	withoutEffectBody(t, CatalogTriggeredBodyKey)
	if _, err := throughJSON(t, g.CaptureSnapshot()).RestoreStrict(); !errors.Is(err, ErrUnknownEffectKey) {
		t.Fatalf("a binary without %s restored the file: err = %v, want ErrUnknownEffectKey", CatalogTriggeredBodyKey, err)
	}
}

// Two instances of one granted bundle are two instances of its
// triggered rows: the second is numbered 1 (ADR 0093 D5's <n>).
func TestMergedGrantRowsAreNumberedPerInstance(t *testing.T) {
	bundle := &CardDef{Triggered: []TriggeredAbility{trigDrainRow(trigDrainKey)}}
	IdentifyCatalogRows(GrantKey("trig-probe/drain"), bundle)
	rows := numberTriggerRowOccurrences(append(append([]TriggeredAbility(nil), bundle.Triggered...), bundle.Triggered...))
	var refs []string
	for _, r := range rows {
		prev := CatalogLookup
		CatalogLookup = func(k string) *CardDef {
			if k == GrantKey("trig-probe/drain") {
				return bundle
			}
			return nil
		}
		ref := TriggeredAbilityRef(r)
		CatalogLookup = prev
		if ref == nil {
			t.Fatal("a granted row was not nameable")
		}
		refs = append(refs, ref.Ref)
	}
	if len(refs) != 2 || refs[0] != "grant:trig-probe/drain:0:0" || refs[1] != "grant:trig-probe/drain:0:1" {
		t.Errorf("refs = %v, want instance 0 then instance 1", refs)
	}
	if bundle.Triggered[0].row.occurrence != 0 {
		t.Error("numbering the merged list wrote to the bundle's own row")
	}
}
