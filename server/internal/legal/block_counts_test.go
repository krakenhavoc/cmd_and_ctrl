package legal_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// block_counts_test.go — #750: the enumerator and the engine agree
// about block COUNTS, not just about pairs.
//
// ADR 0045 §3's invariant is that the verb, the move enumerator and
// the #328 auto-pass signal never disagree, because all three read
// one function. A count broke that: the enumerator offered a lone
// block on a menace attacker, the verb used to take it and undo it
// later, and now the verb refuses it. The three read
// game.BlockOptionsLocked, and these tests state the agreement where
// it would break.

// menaceTable is one menace attacker and `blockers` untapped
// creatures for the defending seat, parked in declare_blockers.
func menaceTable(t *testing.T, blockers int) (*game.Game, uuid.UUID, []uuid.UUID, *game.Player) {
	t.Helper()
	g := newTable(t)
	attackerSeat, defender := g.Seats[0], g.Seats[1]
	clearHand(attackerSeat)
	clearHand(defender)
	attacker := battlefieldCard(g, attackerSeat, game.Card{
		Name: "Menacer", TypeLine: "Creature — Test", Power: 3, Toughness: 3,
		Keywords: []string{"menace"},
	})
	var bs []uuid.UUID
	for i := range blockers {
		bs = append(bs, battlefieldCard(g, defender, game.Card{
			Name:     "Blocker " + string(rune('A'+i)),
			TypeLine: "Creature — Test", Power: 1, Toughness: 4,
		}))
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, defender.ID); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	return g, attacker, bs, defender
}

// A defender with only one creature is offered no block on a menace
// attacker, owes no decision, and is refused if they try anyway.
func TestMenaceWithOneBlockerIsNeverOffered(t *testing.T) {
	g, attacker, bs, defender := menaceTable(t, 1)

	moves := legal.EnumerateFor(g, defender.ID)
	if n := blocksBy(moves, bs[0]); n != 0 {
		t.Fatalf("offered %d blocks the engine would refuse: %v", n, labels(moves))
	}
	if g.SeatOwesBlockDecision(defender.ID) {
		t.Error("the #328 signal holds the window open for a block that cannot be made")
	}
	var refusal *game.BlockRefusedError
	err := g.DeclareBlocker(bs[0], attacker)
	if !errors.As(err, &refusal) || refusal.Reason != game.BlockReasonTooFewBlockers {
		t.Fatalf("declaration refusal = %v, want too_few_blockers", err)
	}
	// Soundness: everything enumerated is accepted.
	dispatchAll(t, g, defender.ID, moves)
}

// With two creatures the block exists — as a GROUP. The enumerator
// offers it as one declare_blockers move rather than as two singles
// the engine would refuse one at a time, and the signal owes the
// decision again.
func TestMenaceWithTwoBlockersIsOfferedAsOneGroupedMove(t *testing.T) {
	g, _, bs, defender := menaceTable(t, 2)

	moves := legal.EnumerateFor(g, defender.ID)
	var groups []legal.Move
	for _, m := range moves {
		if m.Kind != legal.KindBlock {
			continue
		}
		if m.Type != legal.TypeDeclareBlockers {
			t.Errorf("a single block was offered on a menace attacker: %q", m.Label)
			continue
		}
		groups = append(groups, m)
	}
	if len(groups) != 1 {
		t.Fatalf("got %d grouped block moves, want 1: %v", len(groups), labels(moves))
	}
	if groups[0].Source != bs[0] {
		t.Errorf("the group's Source is %v, want the first blocker %v", groups[0].Source, bs[0])
	}
	if groups[0].Label != "Block Menacer with Blocker A and Blocker B" {
		t.Errorf("group label = %q", groups[0].Label)
	}
	if !g.SeatOwesBlockDecision(defender.ID) {
		t.Error("a defender who CAN block a menace attacker must not be auto-passed")
	}
	// The move's params are exactly what performs it (ADR 0033 §1):
	// the server is contractually obliged to accept them.
	dispatchAll(t, g, defender.ID, moves)
}

// The grouped move actually lands: dispatching it blocks the attacker
// with both creatures in one action.
func TestDispatchingTheGroupedMoveBlocksWithBoth(t *testing.T) {
	g, attacker, bs, defender := menaceTable(t, 2)

	var group *legal.Move
	moves := legal.EnumerateFor(g, defender.ID)
	for i := range moves {
		if moves[i].Type == legal.TypeDeclareBlockers {
			group = &moves[i]
			break
		}
	}
	if group == nil {
		t.Fatal("no grouped block move was offered")
	}
	dispatchOne(t, g, defender.ID, *group)

	blocking := map[uuid.UUID]uuid.UUID{}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		blocking[c.InstanceID] = c.BlockingTarget
	}
	for _, b := range bs {
		if blocking[b] != attacker {
			t.Errorf("%v is blocking %v, want the menace attacker %v", b, blocking[b], attacker)
		}
	}
}
