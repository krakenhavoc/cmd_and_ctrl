package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// granted_alternative_cost_cards_test.go — ADR 0118 §3, PR 6 (#2163):
// the four cards that declare the granted-alternative-cost statics,
// each through the real catalog and the real cast path. Per card: the
// offer appears in the picker's list with the source's name in its
// label, claiming it pays the granted price under the strict gate, and
// the offer goes when the source leaves.

const (
	FistOfSunsOracle        = "6fd5e591-fae8-4128-a4d3-a848c8a8ffda"
	JodahArchmageOracle     = "8be4745e-36d8-430f-945e-c8a7fde9b4f6"
	LeylineOfMutationOracle = "caab67eb-65e7-4755-b116-6977e97f0844"
	OmniscienceOracle       = "730e39e6-c61d-48b5-8827-bfd952bf1be7"
)

func GrantedCardOffers(g *game.Game, seat uuid.UUID, card uuid.UUID, zone game.ZoneKind) []*game.AlternativeCost {
	var out []*game.AlternativeCost
	g.ReadSnapshot(func() {
		c, ok := g.LookupCardForEffect(card)
		if !ok {
			return
		}
		out = g.CastOffersForLocked(seat, c, zone, g.CastPermissionForLocked(seat, c, zone))
	})
	return out
}

func OfferWithKey(offers []*game.AlternativeCost, key string) *game.AlternativeCost {
	for _, o := range offers {
		if o != nil && o.Key == key {
			return o
		}
	}
	return nil
}
