package game

import (
	"testing"

	"github.com/google/uuid"
)

// conditional_combat_limits_test.go — #1534, the three shapes ADR
// 0045's #1507 amendment left for their cards (amendment of
// 2026-09-24, Decision 46): the While gate on AttackLimit (Mirri's "as
// long as Mirri is tapped"), the per-permanent scope
// AttackLimitAttackingThis (The Eternal Wanderer), and
// BlockRule.LimitPerDefender (Mirri's "each opponent can't block with
// more than one creature this combat"). Hooks stubbed, as in
// combat_limits_test.go, whose helpers these share; the cards and the
// enumerator are pinned in cards/effects, the bots in aiseat.

// TestAttackLimitWhileGatesTheLimit — untapped, the limit does not
// exist; tapped, it binds, read live; switching it on after two
// attackers are declared unmakes neither; untapped again, it is gone.
func TestAttackLimitWhileGatesTheLimit(t *testing.T) {
	stubCombatLimits(t, []AttackLimit{{
		Scope: AttackLimitAttackingYou, Max: 1,
		While: func(_ *Game, source *Card) bool { return source.Tapped },
	}}, nil)
	g := newFourPlayerActiveGame(t)
	me, mirri := g.Seats[0], g.Seats[1]
	src := pushLimitSource(t, g, mirri, "Mirri")
	var bears []uuid.UUID
	for i := 0; i < 4; i++ {
		bears = append(bears, pushCombatant(t, g, me, "Bear", 2, 2))
	}
	advanceIntoStep(t, g, StepDeclareAttackers)

	for _, b := range bears[:2] {
		if err := g.DeclareAttacker(b, mirri.ID); err != nil {
			t.Fatalf("an untapped source limits nothing: %v", err)
		}
	}
	findCard(g, src).Tapped = true
	if err := g.DeclareAttacker(bears[0], mirri.ID); err != nil {
		t.Errorf("re-declaring a standing attacker under a late limit: %v", err)
	}
	le := attackLimitErr(t, g.DeclareAttacker(bears[2], mirri.ID))
	if le.Defender != mirri.ID || le.Max != 1 || le.Source != src {
		t.Errorf("refusal = %+v", le)
	}
	if n := attackingCount(g); n != 2 {
		t.Errorf("%d attacking, want the 2 declared before the limit switched on", n)
	}
	if err := g.DeclareAttacker(bears[3], g.Seats[2].ID); err != nil {
		t.Fatalf("another opponent: %v", err)
	}
	findCard(g, src).Tapped = false
	if err := g.DeclareAttacker(bears[2], mirri.ID); err != nil {
		t.Errorf("an untapped source still limited the attack: %v", err)
	}
}

// TestAttackLimitAttackingThisCountsOnlyThePermanent — one creature may
// attack the planeswalker; its controller and another planeswalker they
// control may be attacked by any number, and the sentence names the
// planeswalker, the same for every reader.
func TestAttackLimitAttackingThisCountsOnlyThePermanent(t *testing.T) {
	stubCombatLimits(t, []AttackLimit{{Scope: AttackLimitAttackingThis, Max: 1}}, nil)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wanderer := pushPlaneswalkerForTest(g, opp.ID, "The Eternal Wanderer", 5)
	findCard(g, wanderer).OracleID = combatLimitOracle
	other := pushPlaneswalkerForTest(g, opp.ID, "Another Walker", 5)
	var bears []uuid.UUID
	for i := 0; i < 5; i++ {
		bears = append(bears, pushCombatant(t, g, me, "Bear", 2, 2))
	}
	advanceIntoStep(t, g, StepDeclareAttackers)

	if err := g.DeclareAttacker(bears[0], wanderer); err != nil {
		t.Fatalf("one attacker at the Wanderer: %v", err)
	}
	le := attackLimitErr(t, g.DeclareAttacker(bears[1], wanderer))
	if le.Source != wanderer || le.Max != 1 || le.Defender != uuid.Nil || le.Attacker != bears[1] {
		t.Errorf("refusal = %+v", le)
	}
	want := "No more than one creature can attack The Eternal Wanderer each combat."
	for _, viewer := range []uuid.UUID{uuid.Nil, me.ID, opp.ID} {
		if got := le.Sentence(viewer); got != want {
			t.Errorf("viewer %v reads %q, want %q", viewer, got, want)
		}
	}
	for _, b := range bears[1:3] {
		if err := g.DeclareAttacker(b, opp.ID); err != nil {
			t.Fatalf("an attack on the Wanderer's controller was refused: %v", err)
		}
	}
	for _, b := range bears[3:] {
		if err := g.DeclareAttacker(b, other); err != nil {
			t.Fatalf("an attack on another planeswalker was refused: %v", err)
		}
	}
	// Re-pointing an attacker from the player onto the Wanderer raises
	// its count; moving the one away is never refused.
	attackLimitErr(t, g.DeclareAttacker(bears[1], wanderer))
	if err := g.DeclareAttacker(bears[0], opp.ID); err != nil {
		t.Fatalf("lowering the count was refused: %v", err)
	}
	if err := g.DeclareAttacker(bears[1], wanderer); err != nil {
		t.Fatalf("with the slot free the re-point is legal: %v", err)
	}
}

// perDefenderBlockLimit is effects.EachOpponentCantBlockWithMoreThanN
// without importing effects: every blocker not controlled by `you`
// counts toward n, each defending player on their own.
func perDefenderBlockLimit(you uuid.UUID, n int) BlockRule {
	return BlockRule{
		Limit: func(_ *Game, blocker, _ *Card) int {
			if blocker.Controller == you {
				return 0
			}
			return n
		},
		LimitPerDefender: true,
		Label:            "each opponent can't block with more than one creature this combat (Mirri)",
	}
}

// TestBlockLimitPerDefenderCountsEachPlayerOnTheirOwn — four seats.
// Each defending player may block with one creature: one defender's
// block does not use up another's, a second blocker from the same
// defender is refused whichever attacker it names, and the generator
// and the #328 signal agree per seat.
func TestBlockLimitPerDefenderCountsEachPlayerOnTheirOwn(t *testing.T) {
	stubCombatLimits(t, nil, nil)
	g := newFourPlayerActiveGame(t)
	me, p1, p2 := g.Seats[0], g.Seats[1], g.Seats[2]
	atk1 := pushCombatant(t, g, me, "Attacker 1", 3, 3)
	atk1b := pushCombatant(t, g, me, "Attacker 1b", 3, 3)
	atk2 := pushCombatant(t, g, me, "Attacker 2", 3, 3)
	blk1a := pushCombatant(t, g, p1, "P1 Blocker A", 2, 2)
	blk1b := pushCombatant(t, g, p1, "P1 Blocker B", 2, 2)
	blk2a := pushCombatant(t, g, p2, "P2 Blocker A", 2, 2)
	blk2b := pushCombatant(t, g, p2, "P2 Blocker B", 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)
	for _, d := range []AttackDeclaration{{Attacker: atk1, Target: p1.ID}, {Attacker: atk1b, Target: p1.ID}, {Attacker: atk2, Target: p2.ID}} {
		if err := g.DeclareAttacker(d.Attacker, d.Target); err != nil {
			t.Fatal(err)
		}
	}
	g.RegisterTurnScopedBlockRuleLocked(perDefenderBlockLimit(me.ID, 1))
	advanceIntoStep(t, g, StepDeclareBlockers)

	refused := blockLimitRefusal(t, g.DeclareBlockers([]BlockDeclaration{
		{Blocker: blk1a, Attacker: atk1}, {Blocker: blk1b, Attacker: atk1b},
	}))
	if refused.N != 1 || refused.Blocker != blk1a {
		t.Errorf("refusal = %+v", refused)
	}
	if want := "Each opponent can't block with more than one creature this combat (Mirri)."; refused.Sentence(p1.ID) != want {
		t.Errorf("sentence %q, want %q", refused.Sentence(p1.ID), want)
	}
	if err := g.DeclareBlocker(blk1a, atk1); err != nil {
		t.Fatalf("P1's one block: %v", err)
	}
	blockLimitRefusal(t, g.DeclareBlocker(blk1b, atk1b))
	blockLimitRefusal(t, g.DeclareBlocker(blk1b, atk1))
	if opts := g.BlockOptionsLocked(p1.ID, 4); len(opts) != 0 {
		t.Errorf("P1 is offered %d blocks past its one", len(opts))
	}
	if g.SeatOwesBlockDecision(p1.ID) {
		t.Error("the #328 signal holds P1 for a block it cannot make")
	}
	if opts := g.BlockOptionsLocked(p2.ID, 4); len(opts) == 0 {
		t.Fatal("P2 is offered nothing: P1's block used up P2's")
	}
	if err := g.DeclareBlocker(blk2a, atk2); err != nil {
		t.Fatalf("P2's one block: %v", err)
	}
	blockLimitRefusal(t, g.DeclareBlocker(blk2b, atk2))
	if err := g.DeclareBlocker(blk1a, atk1b); err != nil {
		t.Errorf("re-pointing P1's one blocker was refused: %v", err)
	}
}

// TestBlockLimitPerDefenderArrivingLateUnmakesNothing — the raise-the-
// count rule, per group: two blocks from one defender stand when the
// rule arrives, re-pointing one of them is accepted, and a third from
// the same defender is refused.
func TestBlockLimitPerDefenderArrivingLateUnmakesNothing(t *testing.T) {
	stubCombatLimits(t, nil, nil)
	g := newActiveGame(t)
	atkA := pushCombatant(t, g, g.Seats[0], "Attacker A", 2, 2)
	atkB := pushCombatant(t, g, g.Seats[0], "Attacker B", 2, 2)
	blkA := pushCombatant(t, g, g.Seats[1], "Blocker A", 2, 2)
	blkB := pushCombatant(t, g, g.Seats[1], "Blocker B", 2, 2)
	blkC := pushCombatant(t, g, g.Seats[1], "Blocker C", 2, 2)
	declareAttacks(t, g, atkA, atkB)
	if err := g.DeclareBlockers([]BlockDeclaration{
		{Blocker: blkA, Attacker: atkA}, {Blocker: blkB, Attacker: atkA},
	}); err != nil {
		t.Fatalf("no limit yet: %v", err)
	}
	g.RegisterTurnScopedBlockRuleLocked(perDefenderBlockLimit(g.Seats[0].ID, 1))

	if err := g.DeclareBlocker(blkB, atkB); err != nil {
		t.Errorf("re-pointing a standing blocker raises no count: %v", err)
	}
	blockLimitRefusal(t, g.DeclareBlocker(blkC, atkB))
}
