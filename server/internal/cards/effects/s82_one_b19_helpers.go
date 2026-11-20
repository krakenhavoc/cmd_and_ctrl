package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_one_b19_helpers.go — Phyrexia: All Will Be One batch 19 (S82, #5479).

// argentumMasticoreDestroyBody is Argentum Masticore's reflexive "destroy
// target nonland permanent an opponent controls with mana value less than or
// equal to the mana value of the discarded card". The ceiling is the
// discarded card's mana value, fixed when it was discarded (as it left the
// hand) and carried as Params.Amount.
var argentumMasticoreDestroyBody = game.ReflexiveBody("argentum-masticore/destroy",
	SimpleBody(argentumMasticoreDestroy),
	func(_ uuid.UUID, p game.EffectParams) *game.TargetSpec {
		return TargetPermanent("target nonland permanent an opponent controls with mana value less than or equal to the discarded card's mana value",
			Nonland(), OpponentControls(), ManaValueLE(p.Amount))
	})

// capriciousHellraiserNineInGraveyard is the cost gate "if you have nine or
// more cards in your graveyard", read at CR 601.2f off the graveyard as it
// stands.
func CapriciousHellraiserNineInGraveyard(q game.CostQuery) bool {
	if q.Game == nil {
		return false
	}
	return len(GraveyardIDs(q.Game, q.Controller, nil)) >= 9
}
