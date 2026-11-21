package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_gap_b01_helpers.go — S82 GAP batch 1 (#5948).

// GrantedAltCostZeroFromExileOnceYourTurn is "Once during each of your turns,
// you may pay {0} rather than pay the mana cost for a spell you cast from
// exile" (Dragon's Smile). An on-disk identity: never renamed, never reused.
const GrantedAltCostZeroFromExileOnceYourTurn = "granted-zero-from-exile-once-your-turn"

// PayZeroForASpellFromExileOnceEachOfYourTurns is Warped Space's granted offer
// (PayZeroForASpellFromExileOnceEachTurn) limited to the grantor's
// controller's own turn, as Access Maze's is: free, so X is 0 (CR 107.3b),
// only where the printed mana cost could be paid from exile (CR 118.9a), once
// per source object per turn.
func PayZeroForASpellFromExileOnceEachOfYourTurns() game.GrantedAlternativeCost {
	return game.GrantedAlternativeCost{
		Offer: game.AlternativeCost{
			Key:   GrantedAltCostZeroFromExileOnceYourTurn,
			Label: "Pay {0} rather than pay this spell's mana cost",
		},
		Zones:        []game.ZoneKind{game.ZoneExile},
		OnceEachTurn: true,
		YourTurnOnly: true,
	}
}
