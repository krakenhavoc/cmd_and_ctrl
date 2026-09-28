package game

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// multi_block_followups_test.go pins #1715 (ADR 0045 amendment of
// 2026-09-28, Decisions 64-67): block capacity from a RESOLVING effect
// as an ADR 0041 data record (addBlockCapacity / blockAnyNumber), the
// "blocks each attacking creature if able" requirement (blocksEach),
// and the combat-phase defending player a spell's target clause reads.
// The cards are pinned in cards/effects, the enumerator in legal, the
// bots in aiseat.

// capacityOf is BlockCapacity of a battlefield card with fresh layers.
func capacityOf(g *Game, id uuid.UUID) int {
	n := -1
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		n = BlockCapacity(findBattlefieldCard(g, id))
	})
	return n
}

// untilEndOfTurn is this turn's until-end-of-turn duration.
func untilEndOfTurn(g *Game) Duration {
	var d Duration
	g.WithWriteLock(func() { d = g.UntilEndOfTurnDuration() })
	return d
}

// TestTurnScopedBlockCapacityLastsOnlyThisTurn — "can block an
// additional creature this turn": two attackers this turn, one the next
// time it is attacked.
func TestTurnScopedBlockCapacityLastsOnlyThisTurn(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushCombatant(t, g, me, "Bear A", 2, 2)
	b := pushCombatant(t, g, me, "Bear B", 2, 2)
	wall := pushCombatant(t, g, opp, "Wall", 0, 6)
	if got := capacityOf(g, wall); got != 1 {
		t.Fatalf("capacity before the effect = %d, want 1", got)
	}
	registerScopedEffectForTest(t, g, wall, []Mod{AddBlockCapacityMod(1)}, untilEndOfTurn(g))
	declareAttacks(t, g, a, b)
	if got := capacityOf(g, wall); got != 2 {
		t.Fatalf("capacity this turn = %d, want 2", got)
	}
	if err := block(t, g, wall, a, wall, b); err != nil {
		t.Fatalf("blocking two this turn: %v", err)
	}

	advanceToStepOfSeat(t, g, 0, StepDeclareAttackers)
	if got := capacityOf(g, wall); got != 1 {
		t.Fatalf("capacity next turn = %d, want 1 — the record outlived its turn", got)
	}
	for _, id := range []uuid.UUID{a, b} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker next turn: %v", err)
		}
	}
	advanceIntoStep(t, g, StepDeclareBlockers)
	// A capacity-one blocker named twice is RE-POINTED (#1706's
	// sandbox re-declare), so it ends on one attacker.
	if err := block(t, g, wall, a); err != nil {
		t.Fatal(err)
	}
	if err := block(t, g, wall, b); err != nil {
		t.Fatal(err)
	}
	if got := blockedSet(g, wall); !sameIDs(got, []uuid.UUID{b}) {
		t.Fatalf("next turn the wall blocks %v, want only the last attacker named", got)
	}
}

// TestBlockCapacityRecordsAddUpAndAnyNumberBeatsThem — two "+1 this
// turn" records are +2, stacked on a static's +1, and "any number"
// beats every count.
func TestBlockCapacityRecordsAddUpAndAnyNumberBeatsThem(t *testing.T) {
	withMultiBlockStatics(t)
	g := newActiveGame(t)
	opp := g.Seats[1]
	giant := pushMultiBlocker(t, g, opp, "Two-Headed Test", blockTwoOracle, 4, 6)
	d := untilEndOfTurn(g)
	registerScopedEffectForTest(t, g, giant, []Mod{AddBlockCapacityMod(1)}, d)
	registerScopedEffectForTest(t, g, giant, []Mod{AddBlockCapacityMod(2)}, d)
	if got := capacityOf(g, giant); got != 5 {
		t.Fatalf("capacity = %d, want 5 (1 + the static's 1 + 1 + 2)", got)
	}
	registerScopedEffectForTest(t, g, giant, []Mod{BlockAnyNumberMod()}, d)
	if got := capacityOf(g, giant); got != 0 {
		t.Fatalf("capacity = %d, want 0 (any number)", got)
	}
}

func TestAddBlockCapacityModNeedsAPositiveAmount(t *testing.T) {
	g := newActiveGame(t)
	id := pushCombatant(t, g, g.Seats[1], "Wall", 0, 6)
	defer func() {
		if recover() == nil {
			t.Error("an addBlockCapacity mod with no amount registered without a panic")
		}
	}()
	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(id),
			[]Mod{AddBlockCapacityMod(0)}, IndefiniteDuration(), "test")
	})
}

// TestTurnScopedBlockCapacitySurvivesUndoAndSnapshot — the record is
// data, so a clone and a persisted snapshot both carry it, and a
// restore of a file whose capacity mod is malformed is refused.
func TestTurnScopedBlockCapacitySurvivesUndoAndSnapshot(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushCombatant(t, g, me, "Bear A", 2, 2)
	b := pushCombatant(t, g, me, "Bear B", 2, 2)
	wall := pushCombatant(t, g, opp, "Wall", 0, 6)
	saved := g.Clone()
	registerScopedEffectForTest(t, g, wall, []Mod{AddBlockCapacityMod(1), BlockAnyNumberMod()}, untilEndOfTurn(g))
	registerScopedEffectForTest(t, g, a, []Mod{AddBlockRequirementMod(BlockRequirementMustBeBlocked)}, untilEndOfTurn(g))
	declareAttacks(t, g, a, b)

	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("a game holding a capacity record is not a restore point: %+v", snap.Continuations)
	}
	raw, err := json.Marshal(snap)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"addBlockCapacity"`) || !strings.Contains(string(raw), `"blockAnyNumber"`) {
		t.Fatal("the snapshot does not carry the capacity mods by name")
	}
	var decoded GameSnapshot
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	restored, err := decoded.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	if got := capacityOf(restored, wall); got != 0 {
		t.Fatalf("restored capacity = %d, want any number", got)
	}
	if err := block(t, restored, wall, a, wall, b); err != nil {
		t.Fatalf("the restored game refused the multi-block: %v", err)
	}

	// Malformed on disk: refused, not guessed at.
	var bad GameSnapshot
	if err := json.Unmarshal(raw, &bad); err != nil {
		t.Fatal(err)
	}
	for i := range bad.ScopedEffects {
		for j := range bad.ScopedEffects[i].Mods {
			if bad.ScopedEffects[i].Mods[j].Kind == ModAddBlockCapacity {
				bad.ScopedEffects[i].Mods[j].Amount = 0
			}
		}
	}
	if _, err := bad.RestoreStrict(); !errors.Is(err, ErrUnknownEffectKey) {
		t.Fatalf("restoring a capacity mod with no amount: %v, want ErrUnknownEffectKey", err)
	}

	g.RestoreFrom(saved)
	if got := capacityOf(g, wall); got != 1 {
		t.Fatalf("capacity after undo = %d, want 1", got)
	}
}

// TestBlocksEachAttackerMustBlockEveryAttackerItCan — "it blocks each
// attacking creature this turn if able" with "any number": one
// requirement per attacker. Other blockers may join it; leaving an
// attacker it could block is refused at the checkpoint, naming that
// attacker; a flyer it can't block asks nothing.
func TestBlocksEachAttackerMustBlockEveryAttackerItCan(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushCombatant(t, g, me, "Bear A", 2, 2)
	b := pushCombatant(t, g, me, "Bear B", 2, 2)
	c := pushCombatant(t, g, me, "Bear C", 2, 2)
	bird := pushCombatant(t, g, me, "Bird", 1, 1, "flying")
	wall := pushCombatant(t, g, opp, "Wall", 0, 6)
	other := pushCombatant(t, g, opp, "Other", 2, 2)
	registerScopedEffectForTest(t, g, wall,
		[]Mod{BlockAnyNumberMod(), AddBlockRequirementMod(BlockRequirementBlocksEach)}, untilEndOfTurn(g))
	declareAttacks(t, g, a, b, c, bird)

	if err := block(t, g, wall, a); err != nil {
		t.Fatalf("a first block: %v", err)
	}
	if err := block(t, g, other, b); err != nil {
		t.Fatalf("another creature joining on an attacker the wall still owes: %v", err)
	}
	passToDefender(t, g)
	br := blockRequirementErr(t, g.PassPriority())
	if br.Blocker != wall || (br.Attacker != b && br.Attacker != c) {
		t.Errorf("the pass's refusal names %s -> %s, want the wall -> an attacker it left", br.Blocker, br.Attacker)
	}
	if got := br.Sentence(opp.ID); !strings.HasPrefix(got, "Wall must block Bear ") {
		t.Errorf("sentence = %q", got)
	}
	blockRequirementErr(t, g.FinishBlocks(opp.ID))
	if err := block(t, g, wall, b, wall, c); err != nil {
		t.Fatalf("the rest of the attackers: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with every attacker the wall can block blocked (the flyer excepted): %v", err)
	}
	if got := blockedSet(g, wall); !sameIDs(got, []uuid.UUID{a, b, c}) {
		t.Errorf("wall blocks %v, want the three Bears", got)
	}
}

// TestBlocksEachAttackerWithoutAnyNumberFillsItsCapacity — the
// requirement never beats the capacity (a restriction): a creature that
// can block two of three attackers owes two blocks, and two satisfy it.
func TestBlocksEachAttackerWithoutAnyNumberFillsItsCapacity(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushCombatant(t, g, me, "Bear A", 2, 2)
	b := pushCombatant(t, g, me, "Bear B", 2, 2)
	c := pushCombatant(t, g, me, "Bear C", 2, 2)
	wall := pushCombatant(t, g, opp, "Wall", 0, 6)
	registerScopedEffectForTest(t, g, wall,
		[]Mod{AddBlockCapacityMod(1), AddBlockRequirementMod(BlockRequirementBlocksEach)}, untilEndOfTurn(g))
	declareAttacks(t, g, a, b, c)
	if err := block(t, g, wall, a); err != nil {
		t.Fatal(err)
	}
	passToDefender(t, g)
	blockRequirementErr(t, g.PassPriority())
	if err := block(t, g, wall, c); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with the wall full: %v", err)
	}
}

// TestIsDefendingPlayerForEffect — CR 802.2 / 506.2: during the combat
// phase every opponent of the active player, from the beginning of
// combat on; nobody outside it; never the active player.
func TestIsDefendingPlayerForEffect(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	read := func(id uuid.UUID) bool {
		var ok bool
		g.ReadSnapshot(func() { ok = g.IsDefendingPlayerForEffect(id) })
		return ok
	}
	advanceIntoStep(t, g, StepPrecombatMain)
	for _, p := range g.Seats {
		if read(p.ID) {
			t.Errorf("%s is a defending player in the main phase", p.Name)
		}
	}
	advanceIntoStep(t, g, StepBeginCombat)
	for _, p := range g.Seats {
		if want := p.ID != active.ID; read(p.ID) != want {
			t.Errorf("%s at the beginning of combat: defending = %v, want %v", p.Name, !want, want)
		}
	}
	g.WithWriteLock(func() { g.Seats[(g.Turn.ActiveSeat+1)%4].Eliminated = true })
	if read(g.Seats[(g.Turn.ActiveSeat+1)%4].ID) {
		t.Error("an eliminated player is a defending player")
	}
}
