package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// finish_blocks_test.go — #1501. Priority is parked while a defending
// player declares blockers (CR 509.1), so a defender has no pass to say
// "done" with. finish_blocks is the move that says it, and it is what
// keeps a bot defender from holding the table: every seat still
// declaring must always have a move the engine cannot refuse.

// finishMove returns the seat's finish_blocks move, or nil.
func finishMove(moves []legal.Move) *legal.Move {
	for i := range moves {
		if moves[i].Kind == legal.KindFinishBlocks {
			return &moves[i]
		}
	}
	return nil
}

// TestFinishBlocksIsOfferedToADeclaringDefender — the move's shape and
// its lifecycle: offered to the defender only while their declaration
// is pending ("No blocks", then "Done blocking" once a block is
// staged), AlwaysLegal, never offered to the active seat, and gone once
// the declaration is complete. Dispatching it completes the declaration
// and hands the active player priority.
func TestFinishBlocksIsOfferedToADeclaringDefender(t *testing.T) {
	g := newTable(t)
	me := g.Seats[g.Turn.ActiveSeat]
	def := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(me)
	clearHand(def)
	bear := freshCreature(g, me, "Bear")
	wall := freshCreature(g, def, "Wall")
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, def.ID); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if g.Turn.PriorityHolder != game.NoPriority {
		t.Fatalf("setup: priority parked for the declaration, holder %d", g.Turn.PriorityHolder)
	}

	if moves := legal.EnumerateFor(g, me.ID); len(moves) != 0 {
		t.Fatalf("the active seat has nothing to do while priority is parked: %v", labels(moves))
	}
	moves := legal.EnumerateFor(g, def.ID)
	dispatchAll(t, g, def.ID, moves)
	fin := finishMove(moves)
	if fin == nil {
		t.Fatalf("no finish_blocks for the declaring defender: %v", labels(moves))
	}
	if fin.Type != legal.TypeFinishBlocks || fin.Player != def.ID || !fin.AlwaysLegal || fin.Label != "No blocks" {
		t.Errorf("finish move %+v, want an AlwaysLegal finish_blocks for the defender labelled No blocks", *fin)
	}
	if fin.Source != uuid.Nil || len(fin.Params) != 0 {
		t.Errorf("finish move names a card or params: %+v", *fin)
	}
	if countKind(moves, legal.KindPass) != 0 {
		t.Errorf("a pass is offered while nobody holds priority: %v", labels(moves))
	}

	if err := g.DeclareBlocker(wall, bear); err != nil {
		t.Fatal(err)
	}
	moves = legal.EnumerateFor(g, def.ID)
	if fin = finishMove(moves); fin == nil || fin.Label != "Done blocking" {
		t.Fatalf("with a block staged the finish reads Done blocking: %v", labels(moves))
	}
	if err := actions.Dispatch(g, actions.Action{Type: actions.Type(fin.Type), Player: fin.Player, Caller: def.ID, Params: fin.Params}); err != nil {
		t.Fatalf("finish_blocks: %v", err)
	}
	if got := g.BlockDeclarationStatusOf(def.ID); got != game.BlockDeclarationDeclared {
		t.Fatalf("status after the finish %q, want declared", got)
	}
	if g.Turn.PriorityHolder != g.Turn.ActiveSeat {
		t.Fatalf("the active player receives priority after the declaration; holder %d", g.Turn.PriorityHolder)
	}
	if moves := legal.EnumerateFor(g, def.ID); finishMove(moves) != nil || countKind(moves, legal.KindBlock) != 0 {
		t.Errorf("a declared defender is still offered a block or a finish: %v", labels(moves))
	}
	if moves := legal.EnumerateFor(g, me.ID); countKind(moves, legal.KindPass) != 1 {
		t.Errorf("the active player's post-block window offers no pass: %v", labels(moves))
	}
}

// TestEveryDeclaringDefenderCanAlwaysFinish is the stall guard a bot
// depends on: at a four-seat table with three defenders — one with a
// creature, one owing a CR 509.1c requirement, one with nothing to
// block with — every seat still declaring has an AlwaysLegal move, and
// taking the AlwaysLegal move seat by seat ends the declaration with
// priority on the active player. A seat with nothing to block with is
// not waited on at all.
func TestEveryDeclaringDefenderCanAlwaysFinish(t *testing.T) {
	g := newTable(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for _, p := range g.Seats {
		clearHand(p)
	}
	free := g.Seats[(g.Turn.ActiveSeat+1)%4]
	lured := g.Seats[(g.Turn.ActiveSeat+2)%4]
	empty := g.Seats[(g.Turn.ActiveSeat+3)%4]
	var atks []uuid.UUID
	for range 3 {
		atks = append(atks, freshCreature(g, me, "Bear"))
	}
	freshCreature(g, free, "Free Wall")
	freshCreature(g, lured, "Lured Wall")
	lure(t, g, atks[1])
	advanceTo(t, g, game.StepDeclareAttackers)
	for i, p := range []*game.Player{free, lured, empty} {
		if err := g.DeclareAttacker(atks[i], p.ID); err != nil {
			t.Fatal(err)
		}
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if got := g.BlockDeclarationStatusOf(empty.ID); got != game.BlockDeclarationDeclared {
		t.Fatalf("the defender with no creature declared none as the step began: %q", got)
	}
	if g.Turn.PriorityHolder != game.NoPriority {
		t.Fatalf("two defenders declaring: priority parked, holder %d", g.Turn.PriorityHolder)
	}

	for round := 0; round < 8; round++ {
		var next *legal.Move
		var seat uuid.UUID
		for _, p := range g.Seats {
			if g.BlockDeclarationStatusOf(p.ID) != game.BlockDeclarationPending {
				continue
			}
			moves := legal.EnumerateFor(g, p.ID)
			dispatchAll(t, g, p.ID, moves)
			si := -1
			for i, m := range moves {
				if m.AlwaysLegal {
					si = i
					break
				}
			}
			if si < 0 {
				t.Fatalf("round %d: %s is declaring with no AlwaysLegal move — a bot would hold the table: %v", round, p.Name, labels(moves))
			}
			next, seat = &moves[si], p.ID
			break
		}
		if next == nil {
			break
		}
		if err := actions.Dispatch(g, actions.Action{Type: actions.Type(next.Type), Player: next.Player, Caller: seat, Params: next.Params}); err != nil {
			t.Fatalf("round %d: the AlwaysLegal %q was refused: %v", round, next.Label, err)
		}
	}
	for _, p := range []*game.Player{free, lured, empty} {
		if got := g.BlockDeclarationStatusOf(p.ID); got != game.BlockDeclarationDeclared {
			t.Errorf("%s did not finish declaring: %q", p.Name, got)
		}
	}
	if g.Turn.Step != game.StepDeclareBlockers || g.Turn.PriorityHolder != g.Turn.ActiveSeat {
		t.Fatalf("after the last declaration the active player holds priority in the step: %s, holder %d", g.Turn.Step, g.Turn.PriorityHolder)
	}
}
