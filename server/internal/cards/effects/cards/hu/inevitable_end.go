package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inevitable End — Enchantment — Aura {2}{B}:
//
//	"Enchant creature
//	 Enchanted creature has "At the beginning of your upkeep, sacrifice a
//	 creature.""
//
// "Your" is the enchanted creature's controller, who sacrifices a creature
// of their choice (it may be the enchanted one).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e1a5c389-6a56-4da7-b7c4-e26fdec224e5",
		Name:         "Inevitable End",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Grants: []effects.AbilityGrant{{
			Key: inevitableEndGrant,
			Triggered: []game.TriggeredAbility{
				effects.AtYourUpkeep("Inevitable End — sacrifice a creature",
					func(g *game.Game, item *game.StackItem) error {
						g.PlayerSacrificesForEffect(item.SourceCardID, item.Controller,
							effects.SacrificeSpec("a creature", effects.Creature()), "Inevitable End — sacrifice a creature")
						return nil
					}),
			},
			Text: "At the beginning of your upkeep, sacrifice a creature.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(inevitableEndGrant)},
	})
}

// inevitableEndGrant is the granted bundle's key, an on-disk identity:
// never renamed, never reused.
const inevitableEndGrant = "inevitable-end/sacrifice-a-creature"
