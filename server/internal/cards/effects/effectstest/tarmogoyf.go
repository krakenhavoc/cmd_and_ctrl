package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tarmogoyf_test.go covers the S16 sub-PR 5 Tarmogoyf catalog entry
// (Layer 7a CDA): power = distinct card types in all graveyards,
// toughness = power + 1.

const TarmogoyfOracle = "45900b2f-f6a9-4c42-9642-008f3c1cf6dd"

// pushGraveyardCardWithTypeLine seeds a typed card in the named
// player's graveyard. Bumps the layer version directly so the next
// snapshot recomputes — Tarmogoyf's CDA reads from graveyards but
// no zone-move event fires from raw push, so the engine wouldn't
// notice without the explicit bump.
func PushGraveyardCardWithTypeLine(g *game.Game, owner uuid.UUID, name, typeLine string) {
	g.WithWriteLock(func() {
		for _, p := range g.Seats {
			if p.ID != owner {
				continue
			}
			p.Graveyard.PushTop(game.Card{
				InstanceID: uuid.New(),
				Name:       name,
				TypeLine:   typeLine,
				Owner:      owner,
				Controller: owner,
			})
		}
		g.BumpLayerVersionForTest()
	})
}
