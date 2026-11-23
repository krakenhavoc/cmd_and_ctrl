package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Verdant Embrace — Enchantment — Aura {3}{G}{G} (Wilds of Eldraine
// Commander, #5368):
//
//	"Enchant creature
//	 Enchanted creature gets +3/+3 and has "At the beginning of each
//	 upkeep, create a 1/1 green Saproling creature token.""
//
// +3/+3 is a layer 7c pump; the quoted trigger is a granted bundle
// (ADR 0093), so "you" is the enchanted creature's controller and the
// Saprolings are theirs, on every player's upkeep. It ends when the Aura
// leaves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f0bbd988-50b7-476c-9183-dfabd4a95079",
		Name:         "Verdant Embrace",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: verdantEmbraceSaproling,
			Triggered: []game.TriggeredAbility{
				effects.AtEachUpkeep("Verdant Embrace — create a 1/1 green Saproling creature token",
					effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green Saproling"), N: 1})),
			},
			Text: "At the beginning of each upkeep, create a 1/1 green Saproling creature token.",
		}},
		Static: []game.StaticAbility{
			effects.PumpAttached(3, 3),
			effects.GrantAbilitiesToAttached(verdantEmbraceSaproling),
		},
	})
}

const verdantEmbraceSaproling = "verdant-embrace/saproling"
