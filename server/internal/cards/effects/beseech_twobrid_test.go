package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// beseech_twobrid_test.go — #4177, CR 107.4e: Beseech the Queen's
// {2/B}{2/B}{2/B} is paid with {B} or two of anything for each symbol, by
// auto-pay as well as from a floating pool.

func twobridTable(t *testing.T, swamps, plains int) (*game.Game, *game.Player, []uuid.UUID, []uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	var s, p []uuid.UUID
	for i := 0; i < swamps; i++ {
		s = append(s, restrictedBasic(t, g, me, "Swamp"))
	}
	for i := 0; i < plains; i++ {
		p = append(p, restrictedBasic(t, g, me, "Plains"))
	}
	return g, me, s, p
}

func TestBeseechTheQueenIsPaidWithSixLandsOfAnyKind(t *testing.T) {
	g, me, _, plains := twobridTable(t, 0, 6)
	spell := castFromHandForTest(t, g, me, "Beseech the Queen", "Sorcery", "{2/B}{2/B}{2/B}", beseechOracle, autoPay)
	if !g.Stack.Contains(spell) {
		t.Fatal("Beseech the Queen is not on the stack")
	}
	if got := tappedCount(t, g, plains); got != 6 {
		t.Errorf("%d Plains tapped, want 6", got)
	}
}

func TestBeseechTheQueenCostsThreeBlackWhenItCanBePaidThatWay(t *testing.T) {
	g, me, swamps, plains := twobridTable(t, 3, 3)
	castFromHandForTest(t, g, me, "Beseech the Queen", "Sorcery", "{2/B}{2/B}{2/B}", beseechOracle, autoPay)
	if got := tappedCount(t, g, swamps); got != 3 {
		t.Errorf("%d Swamps tapped, want 3", got)
	}
	if got := tappedCount(t, g, plains); got != 0 {
		t.Errorf("%d Plains tapped, want 0: the colour is cheaper", got)
	}
}

func TestBeseechTheQueenMixesBlackAndGeneric(t *testing.T) {
	g, me, swamps, plains := twobridTable(t, 1, 4)
	castFromHandForTest(t, g, me, "Beseech the Queen", "Sorcery", "{2/B}{2/B}{2/B}", beseechOracle, autoPay)
	if tappedCount(t, g, swamps) != 1 || tappedCount(t, g, plains) != 4 {
		t.Errorf("tapped %d Swamps and %d Plains, want 1 and 4", tappedCount(t, g, swamps), tappedCount(t, g, plains))
	}
}

func TestBeseechTheQueenIsNotPaidWithFiveLands(t *testing.T) {
	g, me, _, plains := twobridTable(t, 0, 5)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Beseech the Queen", TypeLine: "Sorcery",
			ManaCost: "{2/B}{2/B}{2/B}", OracleID: beseechOracle, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, autoPay); err == nil {
		t.Fatal("five Plains paid {2/B}{2/B}{2/B}")
	}
	if got := tappedCount(t, g, plains); got != 0 {
		t.Errorf("a refused cast tapped %d lands", got)
	}
}
