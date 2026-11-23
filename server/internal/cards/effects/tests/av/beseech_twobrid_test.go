package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// beseech_twobrid_test.go — #4177, CR 107.4e: Beseech the Queen's
// {2/B}{2/B}{2/B} is paid with {B} or two of anything for each symbol, by
// auto-pay as well as from a floating pool.

func twobridTable(t *testing.T, swamps, plains int) (*game.Game, *game.Player, []uuid.UUID, []uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	var s, p []uuid.UUID
	for i := 0; i < swamps; i++ {
		s = append(s, effectstest.RestrictedBasic(t, g, me, "Swamp"))
	}
	for i := 0; i < plains; i++ {
		p = append(p, effectstest.RestrictedBasic(t, g, me, "Plains"))
	}
	return g, me, s, p
}

func TestBeseechTheQueenIsPaidWithSixLandsOfAnyKind(t *testing.T) {
	g, me, _, plains := twobridTable(t, 0, 6)
	spell := effectstest.CastFromHandForTest(t, g, me, "Beseech the Queen", "Sorcery", "{2/B}{2/B}{2/B}", effectstest.BeseechOracle, effectstest.AutoPay)
	if !g.Stack.Contains(spell) {
		t.Fatal("Beseech the Queen is not on the stack")
	}
	if got := effectstest.TappedCount(t, g, plains); got != 6 {
		t.Errorf("%d Plains tapped, want 6", got)
	}
}

func TestBeseechTheQueenCostsThreeBlackWhenItCanBePaidThatWay(t *testing.T) {
	g, me, swamps, plains := twobridTable(t, 3, 3)
	effectstest.CastFromHandForTest(t, g, me, "Beseech the Queen", "Sorcery", "{2/B}{2/B}{2/B}", effectstest.BeseechOracle, effectstest.AutoPay)
	if got := effectstest.TappedCount(t, g, swamps); got != 3 {
		t.Errorf("%d Swamps tapped, want 3", got)
	}
	if got := effectstest.TappedCount(t, g, plains); got != 0 {
		t.Errorf("%d Plains tapped, want 0: the colour is cheaper", got)
	}
}

func TestBeseechTheQueenMixesBlackAndGeneric(t *testing.T) {
	g, me, swamps, plains := twobridTable(t, 1, 4)
	effectstest.CastFromHandForTest(t, g, me, "Beseech the Queen", "Sorcery", "{2/B}{2/B}{2/B}", effectstest.BeseechOracle, effectstest.AutoPay)
	if effectstest.TappedCount(t, g, swamps) != 1 || effectstest.TappedCount(t, g, plains) != 4 {
		t.Errorf("tapped %d Swamps and %d Plains, want 1 and 4", effectstest.TappedCount(t, g, swamps), effectstest.TappedCount(t, g, plains))
	}
}

func TestBeseechTheQueenIsNotPaidWithFiveLands(t *testing.T) {
	g, me, _, plains := twobridTable(t, 0, 5)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Beseech the Queen", TypeLine: "Sorcery",
			ManaCost: "{2/B}{2/B}{2/B}", OracleID: effectstest.BeseechOracle, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, effectstest.AutoPay); err == nil {
		t.Fatal("five Plains paid {2/B}{2/B}{2/B}")
	}
	if got := effectstest.TappedCount(t, g, plains); got != 0 {
		t.Errorf("a refused cast tapped %d lands", got)
	}
}
