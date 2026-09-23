package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sea Gate Restoration's front face: the hand is counted as the spell
// resolves (the spell itself is on the stack, not in hand), plus one,
// and the controller has no maximum hand size afterwards.
func TestSeaGateRestorationDrawsHandPlusOneAndLiftsTheCap(t *testing.T) {
	g, me, id := handWithMDFC(t, seaGateRestoration)
	for i := 0; i < 3; i++ {
		me.Hand.PushTop(game.NewCard("Filler", me.ID))
	}
	inHand := me.Hand.Size() - 1 // everything but Sea Gate Restoration itself
	library := me.Library.Size()

	floatForTest(g, me, "UUUCCCC")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Sea Gate Restoration: %v", err)
	}
	passPriorityAroundTable(t, g)

	want := inHand + 1
	if drawn := library - me.Library.Size(); drawn != want {
		t.Errorf("drew %d, want %d (a hand of %d, plus one)", drawn, want, inHand)
	}
	if got := me.Hand.Size(); got != inHand+want {
		t.Errorf("hand = %d, want %d", got, inHand+want)
	}
	var cap int
	g.WithWriteLock(func() { cap = g.EffectiveMaxHandSizeLocked(me) })
	if cap != game.NoMaxHandSize {
		t.Errorf("max hand size = %d, want no maximum", cap)
	}
	// Only the caster's cap is lifted.
	for _, p := range g.Seats {
		if p.ID == me.ID {
			continue
		}
		var other int
		g.WithWriteLock(func() { other = g.EffectiveMaxHandSizeLocked(p) })
		if other == game.NoMaxHandSize {
			t.Errorf("an opponent lost their maximum hand size")
		}
	}
}
