package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Darksteel Monolith — Artifact {8} (Commander Masters):
//
//	"Indestructible
//	 Once each turn, you may pay {0} rather than pay the mana cost for a
//	 colorless spell you cast from your hand."
//
// A granted alternative cost (CR 118.9, ADR 0118) like As Foretold's: the
// offer is paying {0}, reaches only spells cast from the hand
// (Zones), only colorless ones (PermissionFilter.ColorlessOnly, read off
// the spell's effective colours), and is claimed once each turn per
// Monolith as the spell becomes cast. It is an alternative cost, so
// additional costs and cost increases still apply.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fdd3482d-7bce-411b-80c4-956a12fc143d",
		Name:            "Darksteel Monolith",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		GrantedAlternativeCosts: []game.GrantedAlternativeCost{{
			Offer: game.AlternativeCost{
				Key:   "granted-zero-colorless-spell-from-hand",
				Label: "Pay {0} rather than pay this spell's mana cost",
			},
			Zones:        []game.ZoneKind{game.ZoneHand},
			Spells:       game.PermissionFilter{ColorlessOnly: true},
			OnceEachTurn: true,
		}},
	})
}
