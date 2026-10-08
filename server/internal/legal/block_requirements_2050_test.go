package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// TestLureThatSparesTheDefenderOwesNothing — #2050: a Lure written with
// the defender as the exempt player ("creatures your opponents control")
// binds none of the defender's creatures, so finishing is offered at
// once and no block is AlwaysLegal-required.
func TestLureThatSparesTheDefenderOwesNothing(t *testing.T) {
	g := newTable(t)
	me := g.Seats[g.Turn.ActiveSeat]
	def := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(me)
	clearHand(def)
	lured := freshCreature(g, me, "Lured Bear")
	freshCreature(g, def, "Wall One")
	freshCreature(g, def, "Wall Two")
	ok := false
	g.WithWriteLock(func() {
		ok = g.RegisterScopedEffectForEffect(uuid.New(), g.PinnedObjectsLocked(lured),
			[]game.Mod{game.AddLureExceptMod(def.ID)}, game.IndefiniteDuration(), "test — Lure sparing the defender")
	})
	if !ok {
		t.Fatal("registered nothing")
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(lured, def.ID); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	moves := legal.EnumerateFor(g, def.ID)
	dispatchAll(t, g, def.ID, moves)
	if countKind(moves, legal.KindFinishBlocks) != 1 {
		t.Errorf("finish not offered though the Lure spares the defender: %v", labels(moves))
	}
	for _, m := range moves {
		if m.AlwaysLegal && m.Kind == legal.KindBlock {
			t.Errorf("a block is required though the Lure spares the defender: %q", m.Label)
		}
	}
}
