package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mirror_shield_hoplite_test.go — #5651: Mirror-Shield Hoplite, and the
// seam it waits on (Event.TargeterKeyword, game.KeywordBackup). Each
// test plays the real catalog entries: the Hoplite beside Guardian
// Scalelord, whose backup 1 grants flying and an attack trigger.

const (
	hopliteOracle      = "23f01235-b5a5-4b2e-a90b-b49fc0c7b18d"
	hopliteScalelord   = "d6c7d7a6-2d38-4c06-ab83-bee9cc18de52"
	hopliteMoogle      = "f718b520-d79d-4b91-9d85-de1f1c90c176"
	hopliteLabel       = "Mirror-Shield Hoplite — copy that ability. You may choose new targets for the copy"
	hopliteTypeLine    = "Creature — Human Soldier"
	hopliteDragonLine  = "Creature — Dragon"
	hopliteScalelordNm = "Guardian Scalelord"
)

type hopliteTable struct {
	g             *game.Game
	me, opp       *game.Player
	hoplite       uuid.UUID
	bear, wolf    uuid.UUID
	oppBear       uuid.UUID
	scalelordCast uuid.UUID
}

func newHopliteTable(t *testing.T) *hopliteTable {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	h := &hopliteTable{g: g, me: me, opp: opp}
	h.hoplite = b12Push(g, me.ID, "Mirror-Shield Hoplite", hopliteTypeLine, hopliteOracle, 2, 2)
	h.bear = b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	h.wolf = b12Creature(g, me.ID, "Wolf", "Creature — Wolf", 2, 2)
	h.oppBear = b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	return h
}

// castBackupAt casts a Guardian Scalelord for the active player and aims
// its backup ability at target. The backup ability is then on the stack,
// with a Hoplite trigger above it if one fired.
func (h *hopliteTable) castBackupAt(t *testing.T, target uuid.UUID) uuid.UUID {
	t.Helper()
	active := h.g.Seats[h.g.Turn.ActiveSeat]
	id := castHoldCreature(t, h.g, hopliteScalelordNm, hopliteDragonLine, hopliteScalelord, 3, 4)
	answerTriggerTargets(t, h.g, active.ID, target)
	h.scalelordCast = id
	return id
}

// resolveUntilCopyPrompt passes priority until the Hoplite's copy asks
// its controller for new targets.
func (h *hopliteTable) resolveUntilCopyPrompt(t *testing.T) *game.PendingChoice {
	t.Helper()
	for i := 0; i < 8; i++ {
		if p := copyPromptFor(h.g, h.me.ID); p != nil {
			return p
		}
		if err := h.g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the Hoplite's copy opened no new-targets prompt")
	return nil
}

func (h *hopliteTable) pickCopyTarget(t *testing.T, target uuid.UUID) {
	t.Helper()
	p := h.resolveUntilCopyPrompt(t)
	if err := h.g.ResolvePickTargets(p.ID, h.me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: target}}); err != nil {
		t.Fatalf("choose the copy's target: %v", err)
	}
}

func hopliteTriggers(g *game.Game, hoplite uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() { n = g.TriggeredThisTurn(hoplite, hopliteLabel) })
	return n
}

func hopliteBecameTargetEvents(g *game.Game) []game.Event {
	var out []game.Event
	g.ReadSnapshot(func() {
		for _, ev := range g.Events {
			if ev.Kind == game.EventBecomesTarget {
				out = append(out, ev)
			}
		}
	})
	return out
}

func TestMirrorShieldHopliteIsFull(t *testing.T) {
	spec, ok := Lookup(hopliteOracle)
	if !ok {
		t.Fatal("Mirror-Shield Hoplite is not registered")
	}
	if spec.Name != "Mirror-Shield Hoplite" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("name %q, completeness %v, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
	}
	h := newHopliteTable(t)
	if !hasEffectiveKeyword(t, h.g, h.hoplite, "vigilance") {
		t.Error("the Hoplite lacks vigilance")
	}
}

// The copy takes a new target, and it is a copy of the Scalelord's
// ability: the Wolf gets the counter and the Scalelord's flying.
func TestMirrorShieldHopliteCopiesTheBackupAbilityAtANewTarget(t *testing.T) {
	h := newHopliteTable(t)
	scalelord := h.castBackupAt(t, h.bear)
	if n := hopliteTriggers(h.g, h.hoplite); n != 1 {
		t.Fatalf("the Hoplite triggered %d times, want 1", n)
	}
	h.pickCopyTarget(t, h.wolf)
	passPriorityAroundTable(t, h.g)
	for name, id := range map[string]uuid.UUID{"Bear": h.bear, "Wolf": h.wolf} {
		if n := countersOn(h.g, id, game.CounterPlusOne); n != 1 {
			t.Errorf("the %s has %d +1/+1 counters, want 1", name, n)
		}
		if !hasEffectiveKeyword(t, h.g, id, "flying") {
			t.Errorf("the %s did not gain the Scalelord's flying", name)
		}
	}
	if hasEffectiveKeyword(t, h.g, h.wolf, "vigilance") {
		t.Error("the copy granted the Hoplite's vigilance: it must grant the backup creature's abilities")
	}
	if countersOn(h.g, h.hoplite, game.CounterPlusOne) != 0 || countersOn(h.g, scalelord, game.CounterPlusOne) != 0 {
		t.Error("a creature that was not targeted got a counter")
	}
}

// "If that's another creature" is asked of the creature with backup, the
// copy's source (CR 707.10b): the Hoplite is another creature and gains
// flying, and the Scalelord targeted by the copy gets the counter only.
func TestMirrorShieldHopliteCopyKeepsTheBackupCreatureAsItsSource(t *testing.T) {
	h := newHopliteTable(t)
	h.castBackupAt(t, h.bear)
	h.pickCopyTarget(t, h.hoplite)
	passPriorityAroundTable(t, h.g)
	if n := countersOn(h.g, h.hoplite, game.CounterPlusOne); n != 1 {
		t.Errorf("the Hoplite has %d counters, want 1", n)
	}
	if !hasEffectiveKeyword(t, h.g, h.hoplite, "flying") {
		t.Error("the Hoplite did not gain flying from the copy")
	}
}

// Keeping the target is allowed (CR 707.10c), and the copy naming my
// creature again does not trigger the Hoplite a second time this turn.
func TestMirrorShieldHopliteMayKeepTheTargetAndTriggersOnce(t *testing.T) {
	h := newHopliteTable(t)
	h.castBackupAt(t, h.bear)
	h.pickCopyTarget(t, h.bear)
	if p := copyPromptFor(h.g, h.me.ID); p != nil {
		t.Fatal("the copy's own targeting triggered the Hoplite again")
	}
	passPriorityAroundTable(t, h.g)
	if n := countersOn(h.g, h.bear, game.CounterPlusOne); n != 2 {
		t.Errorf("the Bear has %d counters, want 2 (the ability and its copy)", n)
	}
	if n := hopliteTriggers(h.g, h.hoplite); n != 1 {
		t.Errorf("the Hoplite triggered %d times this turn, want 1", n)
	}
	evs := hopliteBecameTargetEvents(h.g)
	if len(evs) != 2 {
		t.Fatalf("%d became-target events, want 2 (the ability and its copy)", len(evs))
	}
	for i, ev := range evs {
		if ev.TargeterKind != game.StackItemTriggered || ev.TargeterKeyword != game.KeywordBackup {
			t.Errorf("event %d: kind %q keyword %q, want a triggered backup ability", i, ev.TargeterKind, ev.TargeterKeyword)
		}
	}
}

// A second backup ability the same turn is not copied; one on my next
// turn is.
func TestMirrorShieldHopliteOncePerTurnAndAgainNextTurn(t *testing.T) {
	h := newHopliteTable(t)
	h.castBackupAt(t, h.bear)
	h.pickCopyTarget(t, h.bear)
	passPriorityAroundTable(t, h.g)

	h.castBackupAt(t, h.bear)
	passPriorityAroundTable(t, h.g)
	if copyPromptFor(h.g, h.me.ID) != nil {
		t.Fatal("a second backup ability the same turn was copied")
	}
	if n := countersOn(h.g, h.bear, game.CounterPlusOne); n != 3 {
		t.Fatalf("the Bear has %d counters, want 3 (two from the first, one from the second)", n)
	}

	advanceToUpkeepOf(t, h.g, 1)
	advanceToUpkeepOf(t, h.g, 0)
	h.castBackupAt(t, h.bear)
	h.pickCopyTarget(t, h.bear)
	passPriorityAroundTable(t, h.g)
	if n := countersOn(h.g, h.bear, game.CounterPlusOne); n != 5 {
		t.Errorf("the Bear has %d counters, want 5: the Hoplite copies again on a new turn", n)
	}
}

// A backup ability aimed at a creature I do not control is not copied.
func TestMirrorShieldHopliteIgnoresAnOpposingTarget(t *testing.T) {
	h := newHopliteTable(t)
	h.castBackupAt(t, h.oppBear)
	passPriorityAroundTable(t, h.g)
	if n := hopliteTriggers(h.g, h.hoplite); n != 0 {
		t.Errorf("the Hoplite triggered %d times for an opponent's creature", n)
	}
	if n := countersOn(h.g, h.oppBear, game.CounterPlusOne); n != 1 {
		t.Errorf("their Bear has %d counters, want 1", n)
	}
}

// Whose backup ability it is does not matter: an opponent's, aimed at my
// creature, is copied under my control.
func TestMirrorShieldHopliteCopiesAnOpponentsBackupAbility(t *testing.T) {
	h := newHopliteTable(t)
	advanceToUpkeepOf(t, h.g, 1)
	h.castBackupAt(t, h.bear)
	if n := hopliteTriggers(h.g, h.hoplite); n != 1 {
		t.Fatalf("the Hoplite triggered %d times, want 1", n)
	}
	h.pickCopyTarget(t, h.wolf)
	passPriorityAroundTable(t, h.g)
	for name, id := range map[string]uuid.UUID{"Bear": h.bear, "Wolf": h.wolf} {
		if n := countersOn(h.g, id, game.CounterPlusOne); n != 1 {
			t.Errorf("my %s has %d counters, want 1", name, n)
		}
		if !hasEffectiveKeyword(t, h.g, id, "flying") {
			t.Errorf("my %s did not gain flying", name)
		}
	}
}

// An enters trigger that puts a counter on target creature but is not a
// backup ability does not trigger the Hoplite, and its event names no
// keyword.
func TestMirrorShieldHopliteIgnoresATriggerThatIsNotBackup(t *testing.T) {
	h := newHopliteTable(t)
	castHoldCreature(t, h.g, "Cloudbound Moogle", "Creature — Moogle", hopliteMoogle, 2, 3)
	answerTriggerTargets(t, h.g, h.me.ID, h.bear)
	evs := hopliteBecameTargetEvents(h.g)
	if len(evs) != 1 || evs[0].TargeterKind != game.StackItemTriggered || evs[0].TargeterKeyword != "" {
		t.Fatalf("became-target events %+v, want one triggered event with no keyword", evs)
	}
	passPriorityAroundTable(t, h.g)
	if n := hopliteTriggers(h.g, h.hoplite); n != 0 {
		t.Errorf("the Hoplite triggered %d times for a trigger that is not backup", n)
	}
	if n := countersOn(h.g, h.bear, game.CounterPlusOne); n != 1 {
		t.Errorf("the Bear has %d counters, want 1", n)
	}
}

// The seam across a restore. A table restored with the backup ability's
// target prompt still open stamps the keyword from the restored item's
// catalog row; the event keeps it through a clone and a second restore;
// and the copy made on the restored table resolves as the original does.
func TestMirrorShieldHopliteSeesABackupAbilityAcrossARestore(t *testing.T) {
	h := newHopliteTable(t)
	castHoldCreature(t, h.g, hopliteScalelordNm, hopliteDragonLine, hopliteScalelord, 3, 4)
	h.g = restoreRoundTrip(t, h.g, true)
	h.me = h.g.Seats[0]
	answerTriggerTargets(t, h.g, h.me.ID, h.bear)
	if n := hopliteTriggers(h.g, h.hoplite); n != 1 {
		t.Fatalf("on the restored table the Hoplite triggered %d times, want 1", n)
	}

	// The Hoplite's trigger is on the stack above the backup ability.
	h.g = restoreRoundTrip(t, h.g, true)
	h.me = h.g.Seats[0]
	for name, g := range map[string]*game.Game{"restore": h.g, "clone": h.g.Clone()} {
		evs := hopliteBecameTargetEvents(g)
		if len(evs) != 1 || evs[0].TargeterKeyword != game.KeywordBackup {
			t.Errorf("%s lost the targeter's keyword: %+v", name, evs)
		}
	}
	if n := hopliteTriggers(h.g, h.hoplite); n != 1 {
		t.Fatalf("the restore lost the once-each-turn tally: %d", n)
	}

	h.resolveUntilCopyPrompt(t)
	h.g = strictRestoreWithCopyWalk(t, h.g)
	h.me = h.g.Seats[0]
	h.pickCopyTarget(t, h.wolf)
	passPriorityAroundTable(t, h.g)
	for name, id := range map[string]uuid.UUID{"Bear": h.bear, "Wolf": h.wolf} {
		if n := countersOn(h.g, id, game.CounterPlusOne); n != 1 {
			t.Errorf("the %s has %d counters, want 1", name, n)
		}
		if !hasEffectiveKeyword(t, h.g, id, "flying") {
			t.Errorf("the %s did not gain flying", name)
		}
	}
}

// Backup keeps its ability row on the tile: it has no keyword chip, so
// the row is where the card says it (the Keyword stamp must not hide it).
func TestBackupKeepsItsAbilityRow(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	scalelord := b12Push(g, me.ID, hopliteScalelordNm, hopliteDragonLine, hopliteScalelord, 3, 4)
	c := battlefieldCardCopy(t, g, scalelord)
	rows := game.AbilityRowsOnBattlefield(c)
	triggered := 0
	for _, r := range rows {
		if r.Kind == game.AbilityRowTriggered {
			triggered++
		}
	}
	if triggered != 2 {
		t.Errorf("%d triggered rows on Guardian Scalelord, want 2 (backup and the attack trigger): %+v", triggered, rows)
	}
	stamped := 0
	for _, tr := range game.CatalogTriggers(hopliteScalelord) {
		if tr.Keyword == game.KeywordBackup {
			stamped++
		}
	}
	if stamped != 1 {
		t.Errorf("%d rows of Guardian Scalelord carry the backup keyword, want 1", stamped)
	}
}
