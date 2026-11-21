package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// everyNumber lights every face of the die, so a roll to visit always
// visits the Attraction: the tests below are about what a visit does,
// not about which number came up.
var EveryNumber = []int{1, 2, 3, 4, 5, 6}

// attractionCard is an Attraction card owned by `owner`.
func AttractionCard(owner uuid.UUID, name, oracle string, lights ...int) game.Card {
	return game.Card{
		InstanceID: uuid.New(), Name: name, OracleID: oracle, TypeLine: "Artifact — Attraction",
		Owner: owner, Controller: owner, AstrotoriumBack: true, AttractionLights: lights,
	}
}

// pushAttraction puts an Attraction onto the battlefield under its
// owner's control.
func PushAttraction(g *game.Game, owner uuid.UUID, name, oracle string, lights ...int) uuid.UUID {
	c := AttractionCard(owner, name, oracle, lights...)
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// inMainPhase walks the table to the active player's precombat main
// phase BEFORE any Attraction is on the battlefield, so the CR 717.4
// roll there visits nothing and the test's own roll is the only one.
func InMainPhase(t *testing.T) *game.Game {
	t.Helper()
	g := NewCatalogGame(t)
	AdvanceTo(t, g, game.StepPrecombatMain)
	return g
}

// rollToVisit has `player` roll to visit their Attractions once.
func RollToVisit(t *testing.T, g *game.Game, player uuid.UUID) int {
	t.Helper()
	var result int
	g.WithWriteLock(func() {
		if err := g.RollToVisitForEffect(player, uuid.Nil); err != nil {
			t.Fatalf("roll to visit: %v", err)
		}
		// The result is the roll-to-visit event's (#4307: the visit is the
		// roll's tail).
		for i := len(g.Events) - 1; i >= 0; i-- {
			if ev := g.Events[i]; ev.Kind == game.EventRollToVisit && ev.Actor == player {
				result = ev.Amount
				break
			}
		}
	})
	return result
}
