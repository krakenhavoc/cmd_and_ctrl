package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pushGraveyardCreature seeds a creature card with a given printed
// mana cost into a player's graveyard, and returns its ID.
func PushGraveyardCreature(g *game.Game, owner uuid.UUID, name, manaCost string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		for _, p := range g.Seats {
			if p.ID != owner {
				continue
			}
			p.Graveyard.PushTop(game.Card{
				InstanceID: id,
				Name:       name,
				TypeLine:   "Creature — Bear",
				ManaCost:   manaCost,
				Owner:      owner,
				Controller: owner,
			})
		}
	})
	return id
}
