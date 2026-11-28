package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pendrell Flux — Enchantment — Aura {1}{U} (Urza's Saga, #5624):
//
//	"Enchant creature
//	 Enchanted creature has "At the beginning of your upkeep, sacrifice this
//	 creature unless you pay its mana cost.""
//
// The trigger is granted (ADR 0093): "you" is the creature's controller and
// "its mana cost" is the creature's printed cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fe902db2-3b1d-4266-88eb-307f878fc317",
		Name:         "Pendrell Flux",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Grants: []effects.AbilityGrant{{
			Key: pendrellFluxGrant,
			Triggered: []game.TriggeredAbility{
				effects.AtYourUpkeep("Pendrell Flux — sacrifice this creature unless you pay its mana cost", effects.S82OldB725SacrificeUnlessPayingItsManaCost),
			},
			Text: "At the beginning of your upkeep, sacrifice this creature unless you pay its mana cost.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(pendrellFluxGrant)},
	})
}

const pendrellFluxGrant = "pendrell-flux/upkeep-sacrifice-unless-pay"
