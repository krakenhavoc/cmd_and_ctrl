package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Parasitic Implant — Enchantment — Aura {3}{B} (Jumpstart, #5510):
//
//	"Enchant creature
//	 At the beginning of your upkeep, enchanted creature's controller
//	 sacrifices it and you create a 1/1 colorless Phyrexian Myr artifact
//	 creature token."
//
// On the Aura's controller's upkeep the creature's controller (who may be
// an opponent) sacrifices the enchanted creature, then the Aura's
// controller creates the Myr. The token is made whether or not the creature
// is still there to be sacrificed: the ability has no "if you do".
//
// No simplification.
func init() {
	const label = "Parasitic Implant — enchanted creature's controller sacrifices it and you create a 1/1 Phyrexian Myr"
	effects.Register(effects.Spec{
		OracleID:     "4fc0f58c-5f7f-4a32-8f14-cae696672beb",
		Name:         "Parasitic Implant",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm, Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep(label, effects.Do(
				effects.Instruction(func(ctx *effects.Context) error {
					host := effects.EnchantedCreatureOf(ctx)
					if host == uuid.Nil {
						return nil
					}
					return effects.SacrificePermanent{Target: host}.Apply(ctx)
				}),
				effects.CreateToken{Template: effects.TokenCard("1/1 colorless Phyrexian Myr artifact"), N: 1})),
		},
	})
}
