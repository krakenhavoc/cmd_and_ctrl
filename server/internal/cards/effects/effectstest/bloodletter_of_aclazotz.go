package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// bloodletter_of_aclazotz_test.go covers the card and ADR 0169 (#3388):
// the life that damage costs runs the CR 614 life window, so "lose twice
// that much life" doubles a Bolt and combat damage as well as a drain
// and a life payment. newCatalogGame leaves seat 0 active.

const BloodletterOracle = "469956a2-7cd4-4695-b8f4-c841526f160d"

func PushBloodletter(g *game.Game, owner uuid.UUID) uuid.UUID {
	return B12Push(g, owner, "Bloodletter of Aclazotz", "Creature — Vampire Demon", BloodletterOracle, 2, 4)
}

// lastDamageTo is the last EventDealDamage dealt to a player.
func LastDamageTo(g *game.Game, player uuid.UUID) (game.Event, bool) {
	var last game.Event
	found := false
	for _, ev := range g.Events {
		if ev.Kind == game.EventDealDamage && ev.Target == player {
			last, found = ev, true
		}
	}
	return last, found
}
