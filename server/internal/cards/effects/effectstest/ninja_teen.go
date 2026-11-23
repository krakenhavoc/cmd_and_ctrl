package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pushToGraveyard puts a card into p's graveyard.
func PushToGraveyard(g *game.Game, p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	// A graveyard is public (CR 400.2): every seat knows the card.
	c.KnownBy = map[uuid.UUID]bool{}
	for _, s := range g.Seats {
		c.KnownBy[s.ID] = true
	}
	g.WithWriteLock(func() { p.Graveyard.PushTop(c) })
	return c.InstanceID
}

// graveBear is a vanilla creature card with no way out of a graveyard of
// its own.
func GraveBear() game.Card {
	return game.Card{Name: "Grave Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2}
}

// graveyardOfferKeys lists the claims CastOffersForLocked offers for a
// card in p's graveyard; "" is its mana cost.
func GraveyardOfferKeys(g *game.Game, p *game.Player, card uuid.UUID) []string {
	var keys []string
	g.ReadSnapshot(func() {
		var c game.Card
		for _, gc := range p.Graveyard.Cards {
			if gc.InstanceID == card {
				c = gc
			}
		}
		grant := g.CastPermissionForLocked(p.ID, c, game.ZoneGraveyard)
		for _, ac := range g.CastOffersForLocked(p.ID, c, game.ZoneGraveyard, grant) {
			if ac == nil {
				keys = append(keys, "")
				continue
			}
			keys = append(keys, ac.Key)
		}
	})
	return keys
}
