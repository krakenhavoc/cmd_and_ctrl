package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const CreationOfAvacynOracle = "7f32344f-06d8-4862-bfeb-f90a3e1d7fc2"

// avacynTable casts The Creation of Avacyn for the active seat with one
// tutor target ("Wanted") in the library, resolves chapter I's trigger and
// stops at the search prompt.
func AvacynTable(t *testing.T, wanted game.Card) (g *game.Game, me *game.Player, saga, target uuid.UUID) {
	t.Helper()
	g = NewCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	wanted.Name = "Wanted"
	wanted.Owner, wanted.Controller = me.ID, me.ID
	g.WithWriteLock(func() {
		target = PushLibraryCardForTest(me, wanted)
	})
	saga = CastCatalogSpell(t, g, "The Creation of Avacyn", "Enchantment — Saga", CreationOfAvacynOracle, nil)
	PassPriorityAroundTable(t, g)
	return g, me, saga, target
}

func AvacynNextChapter(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	AdvanceToPrecombatMainOf(t, g, g.Turn.ActiveSeat)
	PassPriorityAroundTable(t, g)
}
