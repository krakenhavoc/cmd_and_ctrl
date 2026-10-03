package legal_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// block_requirements_test.go — #1597, CR 509.1c. The enumerator and the
// verbs agree: every block offered while a requirement is owed is
// accepted, the defending player's pass is withheld exactly while the
// engine refuses it, the required blocks are offered as one AlwaysLegal
// move (so a bot that declines is not left holding the table), and the
// withheld moves really are refused.

func lure(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	ok := false
	g.WithWriteLock(func() {
		ok = g.RegisterScopedEffectForEffect(uuid.New(), g.PinnedObjectsLocked(id),
			[]game.Mod{game.AddBlockRequirementMod(game.BlockRequirementLure)}, game.IndefiniteDuration(), "test — Lure")
	})
	if !ok {
		t.Fatal("registered nothing")
	}
}

// TestLureMovesAgreeWithTheEngine — a Lure'd attacker and a plain one
// against three walls.
func TestLureMovesAgreeWithTheEngine(t *testing.T) {
	g := newTable(t)
	me := g.Seats[g.Turn.ActiveSeat]
	def := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(me)
	clearHand(def)
	lured := freshCreature(g, me, "Lured Bear")
	plain := freshCreature(g, me, "Plain Bear")
	lure(t, g, lured)
	walls := []uuid.UUID{
		freshCreature(g, def, "Wall One"),
		freshCreature(g, def, "Wall Two"),
		freshCreature(g, def, "Wall Three"),
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, a := range []uuid.UUID{lured, plain} {
		if err := g.DeclareAttacker(a, def.ID); err != nil {
			t.Fatal(err)
		}
	}
	advanceTo(t, g, game.StepDeclareBlockers)

	// #1501: priority is parked while the defender declares, so the
	// declaration's "done" is finish_blocks, not a pass.
	moves := legal.EnumerateFor(g, def.ID)
	dispatchAll(t, g, def.ID, moves)
	if n := countKind(moves, legal.KindPass) + countKind(moves, legal.KindFinishBlocks); n != 0 {
		t.Fatalf("pass or finish offered while the Lure is owed: %v", labels(moves))
	}
	var required *legal.Move
	for i, m := range moves {
		if m.Kind != legal.KindBlock {
			continue
		}
		if m.AlwaysLegal {
			if required != nil {
				t.Fatalf("two AlwaysLegal block moves: %v", labels(moves))
			}
			required = &moves[i]
			continue
		}
		// Nothing offered sends a wall at the plain attacker.
		if m.Type == legal.TypeDeclareBlocker {
			var p struct{ Attacker string }
			if err := json.Unmarshal(m.Params, &p); err != nil {
				t.Fatal(err)
			}
			if p.Attacker == plain.String() {
				t.Errorf("offered a wall at the plain attacker, which gives a Lure requirement up: %q", m.Label)
			}
		}
	}
	if required == nil || required.Type != legal.TypeDeclareBlockers {
		t.Fatalf("no AlwaysLegal declare_blockers for the required blocks: %v", labels(moves))
	}
	var set struct {
		Blocks []struct{ Blocker, Attacker string }
	}
	if err := json.Unmarshal(required.Params, &set); err != nil {
		t.Fatal(err)
	}
	if len(set.Blocks) != 3 {
		t.Fatalf("required blocks %+v, want all three walls", set.Blocks)
	}
	for _, b := range set.Blocks {
		if b.Attacker != lured.String() {
			t.Errorf("required block %+v is not on the Lure'd attacker", b)
		}
	}

	// The withheld moves really are refused.
	if err := g.Clone().DeclareBlocker(walls[0], plain); !errors.Is(err, game.ErrIllegalBlock) {
		t.Errorf("engine accepted the withheld block on the plain attacker: %v", err)
	}
	if err := g.Clone().FinishBlocks(def.ID); !errors.Is(err, game.ErrIllegalBlock) {
		t.Errorf("engine accepted the withheld finish: %v", err)
	}

	// Take the required move; the finish comes back.
	decls := make([]game.BlockDeclaration, 0, len(walls))
	for _, w := range walls {
		decls = append(decls, game.BlockDeclaration{Blocker: w, Attacker: lured})
	}
	if err := g.DeclareBlockers(decls); err != nil {
		t.Fatal(err)
	}
	moves = legal.EnumerateFor(g, def.ID)
	dispatchAll(t, g, def.ID, moves)
	if countKind(moves, legal.KindFinishBlocks) != 1 {
		t.Errorf("finish not offered once the Lure is obeyed: %v", labels(moves))
	}
}

// TestNoBlockRequirementLeavesTheMovesAlone — an unaffected table:
// the defender is offered finish_blocks (#1501: priority is parked while
// they declare, so there is no pass), and no move is AlwaysLegal but it.
func TestNoBlockRequirementLeavesTheMovesAlone(t *testing.T) {
	g := newTable(t)
	me := g.Seats[g.Turn.ActiveSeat]
	def := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(me)
	clearHand(def)
	bear := freshCreature(g, me, "Bear")
	freshCreature(g, def, "Wall")
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, def.ID); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	moves := legal.EnumerateFor(g, def.ID)
	if countKind(moves, legal.KindFinishBlocks) != 1 || countKind(moves, legal.KindPass) != 0 {
		t.Fatalf("want finish_blocks and no pass: %v", labels(moves))
	}
	for _, m := range moves {
		if m.AlwaysLegal && m.Kind != legal.KindFinishBlocks {
			t.Errorf("an unaffected table has an AlwaysLegal %q", m.Label)
		}
	}
}
