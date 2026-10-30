package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// World War Hulk's colour clause (#4615): with no land at all, a red creature
// is in the move list under the free-cast promise (and the engine accepts the
// offer), a blue creature and a red sorcery are not.
func TestColouredFreeCastPromiseOffersOnlyTheMatchingSpellsInTheMoveList(t *testing.T) {
	g := newTable(t)
	me := g.Seats[g.Turn.ActiveSeat]
	clearHand(me)
	red := handCard(me, game.Card{Name: "Promise Red Bear", TypeLine: "Creature — Bear", ManaCost: "{4}{R}", Colors: []string{"R"}})
	blue := handCard(me, game.Card{Name: "Promise Blue Bear", TypeLine: "Creature — Bear", ManaCost: "{4}{U}", Colors: []string{"U"}})
	rite := handCard(me, game.Card{Name: "Promise Red Rite", TypeLine: "Sorcery", ManaCost: "{4}{R}", Colors: []string{"R"}})
	advanceTo(t, g, game.StepPrecombatMain)

	if n := len(castMovesFor(legal.EnumerateFor(g, me.ID), red)); n != 0 {
		t.Fatal("setup: a creature costing five with no land was offered")
	}
	grantNextSpell(g, me, game.NextSpellPromise{
		Filter:                game.PermissionFilter{CreatureOnly: true, ColorsAny: "RG"},
		WithoutPayingManaCost: true,
		Text:                  "free",
	})
	moves := castMovesFor(legal.EnumerateFor(g, me.ID), red)
	if len(moves) == 0 {
		t.Fatal("the free-cast promise did not put the red creature in the move list")
	}
	if n := len(castMovesFor(legal.EnumerateFor(g, me.ID), blue)); n != 0 {
		t.Error("a blue creature was offered under a red-or-green promise")
	}
	if n := len(castMovesFor(legal.EnumerateFor(g, me.ID), rite)); n != 0 {
		t.Error("a red sorcery was offered under a creature promise")
	}
	dispatchAll(t, g, me.ID, moves[:1])
}
