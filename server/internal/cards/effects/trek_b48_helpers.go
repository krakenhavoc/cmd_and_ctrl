package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b48_helpers.go — Star Trek Commander batch 48 (#4974).

// trekB48Coven is coven, "as long as you control three or more
// creatures with different powers" (Augur of Autumn). Three creatures
// have different powers only when each power differs from the other
// two, which is the count of distinct powers among creatures you
// control reaching three (rfCreatureCDifferentPowers reads the layered
// power).
func trekB48Coven(g *game.Game, controller, source uuid.UUID) bool {
	return rfCreatureCDifferentPowers(g, controller, source) >= 3
}

// trekB48PlaneswalkersLoyaltyAbilitiesCantBeActivated is The Immortal
// Sun's "Players can't activate planeswalkers' loyalty abilities"
// (CR 606.1): every player, the Sun's controller included, and only a
// loyalty ability whose source is a planeswalker as it stands now.
func trekB48PlaneswalkersLoyaltyAbilitiesCantBeActivated(label string) game.ActivationRestriction {
	return game.ActivationRestriction{
		Label: label,
		Forbids: func(q game.ActivationQuery) bool {
			return q.Ability.Loyalty && q.Card.IsPlaneswalker()
		},
	}
}
