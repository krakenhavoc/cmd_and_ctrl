package legal_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// block_requirements_1684_test.go — #1684's "blocks THAT attacker"
// requirement (Provoke) through the enumerator: every block offered is
// accepted, the defending player's pass is withheld while the provoked
// creature could still block the provoker, the one required move puts
// it on the provoker, and nothing offered sends it at the other
// attacker.

func provokeRecord(t *testing.T, g *game.Game, blocker, attacker uuid.UUID) {
	t.Helper()
	c, ok := g.LookupCardForEffect(attacker)
	if !ok {
		t.Fatal("no attacker")
	}
	ref := game.ObjectRef{ID: attacker, Epoch: c.ObjectEpoch}
	registered := false
	g.WithWriteLock(func() {
		registered = g.RegisterScopedEffectForEffect(uuid.New(), g.PinnedObjectsLocked(blocker),
			[]game.Mod{game.BlocksAttackerMod(ref)}, game.IndefiniteDuration(), "test — Provoke")
	})
	if !registered {
		t.Fatal("registered nothing")
	}
}

func TestProvokeMovesAgreeWithTheEngine(t *testing.T) {
	g := newTable(t)
	me := g.Seats[g.Turn.ActiveSeat]
	def := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(me)
	clearHand(def)
	provoker := freshCreature(g, me, "Provoker")
	plain := freshCreature(g, me, "Plain Bear")
	provoked := freshCreature(g, def, "Provoked Bear")
	freshCreature(g, def, "Free Bear")
	provokeRecord(t, g, provoked, provoker)
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, a := range []uuid.UUID{provoker, plain} {
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
		t.Fatalf("pass or finish offered while the provoked block is owed: %v", labels(moves))
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
		if m.Type == legal.TypeDeclareBlocker {
			var p struct{ Blocker, Attacker string }
			if err := json.Unmarshal(m.Params, &p); err != nil {
				t.Fatal(err)
			}
			if p.Blocker == provoked.String() && p.Attacker != provoker.String() {
				t.Errorf("offered the provoked creature at another attacker: %q", m.Label)
			}
		}
	}
	if required == nil {
		t.Fatalf("no AlwaysLegal block for the provoked creature: %v", labels(moves))
	}
	var one struct{ Blocker, Attacker string }
	if err := json.Unmarshal(required.Params, &one); err != nil {
		t.Fatal(err)
	}
	if required.Type != legal.TypeDeclareBlocker || one.Blocker != provoked.String() || one.Attacker != provoker.String() {
		t.Fatalf("required move %s %s, want the provoked creature on the provoker", required.Type, required.Params)
	}
	if err := g.Clone().DeclareBlocker(provoked, plain); !errors.Is(err, game.ErrIllegalBlock) {
		t.Errorf("engine accepted the withheld block on the other attacker: %v", err)
	}
	if err := g.Clone().FinishBlocks(def.ID); !errors.Is(err, game.ErrIllegalBlock) {
		t.Errorf("engine accepted the withheld finish: %v", err)
	}
	if err := g.DeclareBlocker(provoked, provoker); err != nil {
		t.Fatal(err)
	}
	moves = legal.EnumerateFor(g, def.ID)
	dispatchAll(t, g, def.ID, moves)
	if countKind(moves, legal.KindFinishBlocks) != 1 {
		t.Errorf("finish not offered once the provoked block is made: %v", labels(moves))
	}
}
