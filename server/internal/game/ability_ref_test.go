package game

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// ability_ref_test.go — ADR 0041 phase 3 tier 4, first slice (P9,
// #1497): an activated ability's stack item names its catalog row, so a
// table with one waiting is a restore point; restore re-derives the
// ability with the owner's Q2 name check, and restores a row it cannot
// find as a manual item (Q3). Every test swaps CatalogLookup between the
// capture and the restore, which is what a deploy does.

const refOracle = "fixture-ability-ref"

const (
	refGainLabel = "{T}: You gain 1 life."
	refPingLabel = "{0}: Target player loses 1 life."
)

func refGainRow() ActivatedAbilityShape {
	return ActivatedAbilityShape{
		Label: refGainLabel,
		Effect: func(g *Game, item *StackItem) error {
			return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, 1)
		},
	}
}

func refPingRow(label string) ActivatedAbilityShape {
	return ActivatedAbilityShape{
		Label:   label,
		Targets: &TargetSpec{Mode: "player", Label: "target player", Players: true, Min: 1, Max: 1},
		Effect: func(g *Game, item *StackItem) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Targets[0].ID, -1)
		},
	}
}

// withRefCatalog installs a one-card catalog whose activated rows are
// `rows`, for the rest of the test (or until the next call).
func withRefCatalog(t *testing.T, rows ...ActivatedAbilityShape) {
	t.Helper()
	withCatalog(t, refOracle, &CardDef{Activated: rows})
}

// activateRefPing puts the probe on the battlefield and activates its
// ping at `index`, targeting seat 1. It returns the probe and the item.
func activateRefPing(t *testing.T, g *Game, index int) (uuid.UUID, *StackItem) {
	t.Helper()
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	probe := pushTypedTestCard(g, Card{
		Name: "Ref Probe", TypeLine: "Artifact", OracleID: refOracle,
		Owner: me, Controller: me,
	})
	if err := g.ActivateCatalogAbility(me, probe, index, ActivateAbilityParams{
		Targets: []TargetRef{{Kind: TargetPlayer, ID: opp}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	var item *StackItem
	g.ReadSnapshot(func() {
		for _, it := range g.StackMeta {
			if it.SourceCardID == probe {
				item = it
			}
		}
	})
	if item == nil {
		t.Fatal("the activation put no item on the stack")
	}
	return probe, item
}

// restoredStackItem is the one ability item on a restored game's stack.
func restoredStackItem(t *testing.T, g *Game) *StackItem {
	t.Helper()
	var out *StackItem
	g.ReadSnapshot(func() {
		for _, it := range g.StackMeta {
			if it.Kind == StackItemActivated {
				out = it
			}
		}
	})
	if out == nil {
		t.Fatal("the restored game has no ability on its stack")
	}
	return out
}

func resolveRefAbility(g *Game) {
	g.WithWriteLock(func() { g.resolveTopAbilityLocked() })
}

func TestActivatedAbilityOnTheStackIsARestorePoint(t *testing.T) {
	withRefCatalog(t, refGainRow(), refPingRow(refPingLabel))
	g := newActiveGame(t)
	_, item := activateRefPing(t, g, 1)

	want := AbilityRef{Key: refOracle, Slot: AbilitySlotActivated, Ref: "own:1", Name: refPingLabel}
	if item.Body != CatalogActivatedBodyKey || item.Params.Ability == nil || *item.Params.Ability != want {
		t.Fatalf("item stamp = body %q ref %+v, want %q %+v", item.Body, item.Params.Ability, CatalogActivatedBodyKey, want)
	}
	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("a stamped activated ability on the stack blocks the restore point: %+v", snap.Continuations)
	}
	restored, err := throughJSON(t, snap).RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	back := restoredStackItem(t, restored)
	if back.Effect == nil || back.targetSpec == nil {
		t.Fatalf("restore did not re-derive the row: effect %v, target spec %v", back.Effect != nil, back.targetSpec != nil)
	}
	opp := restored.Seats[1]
	before := opp.Life
	resolveRefAbility(restored)
	if opp.Life != before-1 {
		t.Errorf("the restored ping resolved to life %d, want %d", opp.Life, before-1)
	}
	// The restored game is a restore point again, and round-trips.
	if again := restored.CaptureSnapshot(); !again.Restorable() {
		t.Errorf("the restored game is not a restore point: %+v", again.Continuations)
	}
}

// An ability copy carries the stamp too: it is the same row.
func TestAbilityCopyOfAStampedItemKeepsItsRef(t *testing.T) {
	withRefCatalog(t, refGainRow(), refPingRow(refPingLabel))
	g := newActiveGame(t)
	_, item := activateRefPing(t, g, 1)
	g.WithWriteLock(func() { g.createAbilityCopyLocked(item, g.Seats[0].ID, item.Targets) })
	n := 0
	g.ReadSnapshot(func() {
		for _, it := range g.StackMeta {
			if it.IsCopy {
				n++
				if it.Body != CatalogActivatedBodyKey || it.Params.Ability == nil || it.Params.Ability == item.Params.Ability {
					t.Errorf("copy stamp = %q %+v, want its own copy of the original's ref", it.Body, it.Params.Ability)
				}
			}
		}
	})
	if n != 1 {
		t.Fatalf("%d copies, want 1", n)
	}
	if snap := g.CaptureSnapshot(); !snap.Restorable() {
		t.Errorf("the copy blocks the restore point: %+v", snap.Continuations)
	}
}

// Q2: a deploy reordered the card's abilities. The ref no longer names
// the ping, but exactly one row carries its label, and that is the row.
func TestStackAbilityRefFollowsItsLabelAfterAReorder(t *testing.T) {
	withRefCatalog(t, refGainRow(), refPingRow(refPingLabel))
	g := newActiveGame(t)
	probe, _ := activateRefPing(t, g, 1)
	snap := throughJSON(t, g.CaptureSnapshot())

	withRefCatalog(t, refPingRow(refPingLabel), refGainRow())
	if lost := snap.LostStackAbilities(); len(lost) != 0 {
		t.Fatalf("a reordered row is reported lost: %+v", lost)
	}
	restored, err := snap.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	back := restoredStackItem(t, restored)
	if back.Params.Ability == nil || back.Params.Ability.Ref != "own:0" {
		t.Fatalf("ref after the reorder = %+v, want own:0", back.Params.Ability)
	}
	opp := restored.Seats[1]
	before := opp.Life
	resolveRefAbility(restored)
	if opp.Life != before-1 {
		t.Errorf("the ping resolved to life %d, want %d — the gain row ran instead?", opp.Life, before-1)
	}
	restored.ReadSnapshot(func() {
		if c := restored.findCardByIDLocked(probe); c == nil || c.AbilitiesLostOnRestore {
			t.Error("a card whose ability was found is flagged lost")
		}
	})
}

// Q3: the row is gone — reworded — and nothing else carries the label.
// The game restores; the item stays on the stack as a manual one; the
// card is flagged; the boot report names it.
func TestStackAbilityThatNoRowAnswersRestoresAsAManualItem(t *testing.T) {
	withRefCatalog(t, refGainRow(), refPingRow(refPingLabel))
	g := newActiveGame(t)
	probe, item := activateRefPing(t, g, 1)
	itemID := item.ID
	snap := throughJSON(t, g.CaptureSnapshot())

	withRefCatalog(t, refGainRow(), refPingRow("{0}: Target player loses 1 life. (reworded)"))
	lost := snap.LostStackAbilities()
	if len(lost) != 1 || lost[0].ItemID != itemID || lost[0].CardID != probe ||
		lost[0].Card != "Ref Probe" || lost[0].Ref.Ref != "own:1" || lost[0].Ref.Name != refPingLabel {
		t.Fatalf("LostStackAbilities = %+v, want the probe's ping", lost)
	}
	restored, err := snap.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict refused the table over one ability: %v", err)
	}
	back := restoredStackItem(t, restored)
	if back.ID != itemID || back.Effect != nil || back.Body != "" || back.Params.Ability != nil ||
		back.targetSpec != nil || back.modeSpec != nil {
		t.Errorf("the lost item came back as %+v, want a manual item with nothing to run", back)
	}
	if len(back.Targets) != 1 || back.Label != refPingLabel {
		t.Errorf("the manual item lost its announcement: targets %v label %q", back.Targets, back.Label)
	}
	restored.ReadSnapshot(func() {
		if c := restored.findCardByIDLocked(probe); c == nil || !c.AbilitiesLostOnRestore {
			t.Error("the source card is not flagged AbilitiesLostOnRestore")
		}
	})
	opp := restored.Seats[1]
	before := opp.Life
	resolveRefAbility(restored)
	if opp.Life != before {
		t.Errorf("the manual item did something: life %d -> %d", before, opp.Life)
	}
	// A recapture is a restore point and does not report it again.
	again := restored.CaptureSnapshot()
	if !again.Restorable() || len(again.LostStackAbilities()) != 0 {
		t.Errorf("the recapture: restorable %v, lost %+v", again.Restorable(), again.LostStackAbilities())
	}
}

// Q3 includes "more than one": two rows with the label is no answer.
func TestStackAbilityRefWithAnAmbiguousLabelIsLost(t *testing.T) {
	withRefCatalog(t, refGainRow(), refPingRow(refPingLabel))
	g := newActiveGame(t)
	activateRefPing(t, g, 1)
	snap := throughJSON(t, g.CaptureSnapshot())

	withRefCatalog(t, refPingRow(refPingLabel), refPingRow(refPingLabel), refGainRow())
	// own:1 is now the SECOND ping, which carries the label: a match.
	if lost := snap.LostStackAbilities(); len(lost) != 0 {
		t.Fatalf("the row at the ref matches, but it is reported lost: %+v", lost)
	}
	withRefCatalog(t, refGainRow(), refGainRow(), refPingRow(refPingLabel), refPingRow(refPingLabel))
	if lost := snap.LostStackAbilities(); len(lost) != 1 {
		t.Fatalf("two rows carry the label and neither is at the ref: lost = %+v, want 1", lost)
	}
}

// A granted row names its bundle, and restores through it.
func TestGrantedActivatedAbilityOnTheStackRestores(t *testing.T) {
	fx := stubGrantedAbilityCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	bear := gaBear(g, me)
	gaPush(g, me, "Pinger", "Enchantment", gaPingerOracle)
	g.BumpLayerVersionForTest()
	_, origins := ActivatedAbilitiesWithOrigins(layeredBattlefieldCard(t, g, bear))
	if err := g.ActivateCatalogAbility(me, bear, 0, ActivateAbilityParams{Ref: origins.Ref(0)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("a granted ability on the stack blocks the restore point: %+v", snap.Continuations)
	}
	var ref *AbilityRef
	for _, it := range snap.StackMeta {
		if it.Params != nil {
			ref = it.Params.Ability
		}
	}
	want := AbilityRef{Key: GrantKey(gaPingBundle), Slot: AbilitySlotActivated, Ref: GrantedAbilityRef(gaPingBundle, 0, 0), Name: "{1}: ping"}
	if ref == nil || *ref != want {
		t.Fatalf("granted stamp = %+v, want %+v", ref, want)
	}
	restored, err := throughJSON(t, snap).RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	resolveRefAbility(restored)
	if fx.pinged != 1 {
		t.Errorf("the restored granted ability ran %d times, want 1", fx.pinged)
	}
}

// An ability with no catalog identity — carried on the instance — is
// not stamped, and still blocks the restore point as it always did.
func TestInstanceAbilityIsNotStamped(t *testing.T) {
	withRefCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	probe := pushTypedTestCard(g, Card{
		Name: "Contraption", TypeLine: "Artifact", Owner: me, Controller: me,
		ActivatedAbilities: []ActivatedAbilityShape{refGainRow()},
	})
	gaSettle(g, probe)
	if err := g.ActivateCatalogAbility(me, probe, 0, ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	g.ReadSnapshot(func() {
		for _, it := range g.StackMeta {
			if it.Body != "" || it.Params.Ability != nil {
				t.Errorf("an instance ability was stamped: %q %+v", it.Body, it.Params.Ability)
			}
		}
	})
	if snap := g.CaptureSnapshot(); snap.Continuations.StackEffects != 1 {
		t.Errorf("census = %+v, want the unkeyed item counted once in StackEffects", snap.Continuations)
	}
}

// --- the refusals (ADR 0041 P10) --------------------------------------

// stampedSnapshotJSON is a capture holding one stamped activated
// ability, as generic JSON a test can edit.
func stampedSnapshotJSON(t *testing.T) map[string]any {
	t.Helper()
	withRefCatalog(t, refGainRow(), refPingRow(refPingLabel))
	g := newActiveGame(t)
	activateRefPing(t, g, 1)
	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func restoreEdited(t *testing.T, m map[string]any) error {
	t.Helper()
	raw, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	var s GameSnapshot
	if err := json.Unmarshal(raw, &s); err != nil {
		t.Fatal(err)
	}
	_, err = s.RestoreStrict()
	return err
}

func firstStackItemJSON(t *testing.T, m map[string]any) map[string]any {
	t.Helper()
	list, _ := m["stackMeta"].([]any)
	if len(list) != 1 {
		t.Fatalf("want one stack item in the capture, have %d", len(list))
	}
	return list[0].(map[string]any)
}

func abilityJSON(t *testing.T, it map[string]any) map[string]any {
	t.Helper()
	params, _ := it["params"].(map[string]any)
	ab, _ := params["ability"].(map[string]any)
	if ab == nil {
		t.Fatalf("the stack item carries no ability ref: %v", it)
	}
	return ab
}

// withoutEffectBody is a binary from before a body was registered: the
// same registry, less one key. P10's claim is that such a binary — every
// v7 binary before this slice — refuses a file naming it.
func withoutEffectBody(t *testing.T, key string) {
	t.Helper()
	effectRegistryMu.Lock()
	fn, ok := effectBodies[key]
	delete(effectBodies, key)
	effectRegistryMu.Unlock()
	if !ok {
		t.Fatalf("%s is not registered", key)
	}
	t.Cleanup(func() {
		effectRegistryMu.Lock()
		effectBodies[key] = fn
		effectRegistryMu.Unlock()
	})
}

func TestABinaryWithoutTheCatalogBodyRefusesAStampedAbility(t *testing.T) {
	m := stampedSnapshotJSON(t)
	if err := restoreEdited(t, m); err != nil {
		t.Fatalf("control: this binary refuses its own file: %v", err)
	}
	withoutEffectBody(t, CatalogActivatedBodyKey)
	if err := restoreEdited(t, m); !errors.Is(err, ErrUnknownEffectKey) {
		t.Fatalf("a binary without %s restored the file: err = %v, want ErrUnknownEffectKey", CatalogActivatedBodyKey, err)
	}
}

func TestStackItemRefusals(t *testing.T) {
	cases := []struct {
		name string
		edit func(it map[string]any)
		want string
	}{
		{"an unknown field on the stack item", func(it map[string]any) { it["futureField"] = true }, `unknown field "futureField" on a stack item`},
		{"an unknown field on the ability ref", func(it map[string]any) {
			abilityJSON(t, it)["future"] = "x"
		}, `unknown field "future" on a stack item's params ability`},
		{"an unknown slot", func(it map[string]any) { abilityJSON(t, it)["slot"] = "static" }, `ability-ref slot "static"`},
		{"the triggered slot, which this binary does not read yet", func(it map[string]any) { abilityJSON(t, it)["slot"] = "triggered" }, `ability-ref slot "triggered"`},
		{"a ref outside the grammar", func(it map[string]any) { abilityJSON(t, it)["ref"] = "land:G" }, `ability ref "land:G"`},
		{"a grant ref naming another bundle than its key", func(it map[string]any) {
			abilityJSON(t, it)["ref"] = "grant:somebody/else:0:0"
		}, `ability ref "grant:somebody/else:0:0"`},
		{"the body with no ref", func(it map[string]any) { delete(it, "params") }, "stack item with no ability ref"},
		{"a ref under a body that does not read one", func(it map[string]any) { it["body"] = "draw/one-card" }, "does not read one"},
		{"the triggered body, which is 4-2's", func(it map[string]any) { it["body"] = "catalog/triggered" }, "stack-item body catalog/triggered"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := stampedSnapshotJSON(t)
			tc.edit(firstStackItemJSON(t, m))
			err := restoreEdited(t, m)
			if !errors.Is(err, ErrUnknownEffectKey) {
				t.Fatalf("err = %v, want ErrUnknownEffectKey", err)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("err = %v, want it to name %q", err, tc.want)
			}
		})
	}
}

// The whole-record refusal is asked of the CURRENT schema only: in an
// older file an unknown key can only be one a later bump removed.
func TestStackItemFieldRefusalIsForTheCurrentSchema(t *testing.T) {
	raw := []byte(`{"stackMeta":[{"id":"00000000-0000-0000-0000-000000000001","removedLongAgo":1}]}`)
	older, err := unknownScopedEffectFields(raw, SnapshotSchemaVersion-1)
	if err != nil {
		t.Fatal(err)
	}
	if len(older) != 0 {
		t.Errorf("an older file's unknown stack-item key is refused: %v", older)
	}
	current, err := unknownScopedEffectFields(raw, SnapshotSchemaVersion)
	if err != nil {
		t.Fatal(err)
	}
	if len(current) != 1 {
		t.Errorf("a current file's unknown stack-item key: %v, want one refusal", current)
	}
}

// An ability ref belongs to a stack item; a delayed trigger naming one
// is a shape no binary writes.
func TestDelayedTriggerWithAnAbilityRefIsRefused(t *testing.T) {
	s := &GameSnapshot{Schema: SnapshotSchemaVersion, DelayedTriggers: []delayedTriggerSnapshot{{
		Body:   "draw/one-card",
		Params: &EffectParams{Ability: &AbilityRef{Key: refOracle, Slot: AbilitySlotActivated, Ref: "own:0"}},
	}}}
	if err := s.checkEffectKeys(); !errors.Is(err, ErrUnknownEffectKey) {
		t.Errorf("err = %v, want ErrUnknownEffectKey", err)
	}
}

// P10's second rule, for lastKnownStack: a binary from before it drops
// the key, which is what it does with files it writes itself — a file
// without it still restores, and restores the record empty.
func TestLastKnownStackIsAdditive(t *testing.T) {
	g := newActiveGame(t)
	id := pushLKISpell(t, g, g.Seats[0].ID, "Opt")
	counterForTest(t, g, id)
	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	if _, ok := m["lastKnownStack"]; !ok {
		t.Fatal("the capture carries no lastKnownStack")
	}
	delete(m, "lastKnownStack")
	raw, _ = json.Marshal(m)
	var s GameSnapshot
	if err := json.Unmarshal(raw, &s); err != nil {
		t.Fatal(err)
	}
	restored, err := s.RestoreStrict()
	if err != nil {
		t.Fatalf("a file without lastKnownStack: %v", err)
	}
	if len(restored.lastKnownStack) != 0 {
		t.Errorf("restored %d records from nothing", len(restored.lastKnownStack))
	}
}
