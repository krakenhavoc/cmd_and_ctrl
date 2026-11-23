package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// flashbackOffers is the cast offers for a card in seat's graveyard, by
// claim key, each with its mana cost.
func FlashbackOffers(t *testing.T, g *game.Game, seat, id uuid.UUID) map[string]string {
	t.Helper()
	out := map[string]string{}
	g.ReadSnapshot(func() {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			t.Fatal("the card is gone")
		}
		for _, ac := range g.CastOffersForLocked(seat, c, game.ZoneGraveyard, g.CastPermissionForLocked(seat, c, game.ZoneGraveyard)) {
			if ac == nil {
				out[""] = c.ManaCost
				continue
			}
			out[ac.Key] = ac.ManaCost
		}
	})
	return out
}

// setManaCost gives a seeded card its printed mana cost, which the
// catalog game does not fill in.
func SetManaCost(g *game.Game, seat *game.Player, id uuid.UUID, cost string) {
	g.WithWriteLock(func() {
		for i := range seat.Graveyard.Cards {
			if seat.Graveyard.Cards[i].InstanceID == id {
				seat.Graveyard.Cards[i].ManaCost = cost
			}
		}
	})
}
