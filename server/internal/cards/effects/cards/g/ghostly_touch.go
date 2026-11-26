package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ghostly Touch — {1}{U} Aura: "Enchant creature. Enchanted creature has
// "Whenever this creature attacks, you may tap or untap target permanent.""
//
// No simplification.
func init() {
	const grant = "ghostly-touch/attack"
	const text = "Whenever this creature attacks, you may tap or untap target permanent."
	effects.Register(effects.Spec{
		OracleID:     "02d25219-888d-4032-9e9e-8c6f879a9084",
		Name:         "Ghostly Touch",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Triggered: []game.TriggeredAbility{
				effects.Optional(
					effects.Targeting(
						effects.WheneverThisAttacks("Ghostly Touch — tap or untap target permanent", effects.B38TapOrUntapTarget("Ghostly Touch")),
						effects.TargetPermanent("target permanent", effects.Permanent())),
					"Ghostly Touch — tap or untap target permanent?"),
			},
			Text: text,
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(grant)},
	})
}
