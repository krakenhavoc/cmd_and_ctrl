package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tannuk, Steadfast Second — Legendary Creature — Kavu Pilot {2}{R}{R},
// 3/5 (Edge of Eternities, #4461):
//
//	"Other creatures you control have haste.
//	 Artifact cards and red creature cards in your hand have warp
//	 {2}{R}."
//
// Aggressive Mammoth's keyword grant for haste. The warp line is a granted
// alternative cost (WarpForArtifactAndRedCreatureCardsInHand): derived
// from the battlefield on every cast query, so it lasts exactly as long as
// Tannuk is under his controller's control, reaches only the cards in the
// hand (an artifact card, or a creature card that is red), and exiles the
// permanent at the next end step with the "cast it from exile on a later
// turn" permission like printed warp. The cast is a warp cast: it counts
// for void and for "if you cast it from warp" riders.
//
// Seam: game.GrantedAltCostWarp, a granted offer that may carry warp's
// exile clause.
func init() {
	Register(Spec{
		OracleID:                "b5ae8016-92ed-4ae8-83a8-65068defc616",
		Name:                    "Tannuk, Steadfast Second",
		Completeness:            CompletenessFull,
		GrantedAlternativeCosts: []game.GrantedAlternativeCost{WarpForArtifactAndRedCreatureCardsInHand("{2}{R}")},
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Others: true, YoursOnly: true}, "haste"),
		},
	})
}
