package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Presence of Gond — Enchantment — Aura {2}{G} (Jumpstart 2022):
//
//	"Enchant creature
//	 Enchanted creature has "{T}: Create a 1/1 green Elf Warrior creature
//	 token.""
//
// Splinter Twin's granted activated ability: the creature, not the Aura, is
// the source, so its controller makes the token, it needs the creature to be
// untapped and free of summoning sickness (CR 302.6), and it is lost with the
// Aura.
//
// No simplification.
func init() {
	const grant = "presence-of-gond/elf-warrior"
	const text = "{T}: Create a 1/1 green Elf Warrior creature token."
	effects.Register(effects.Spec{
		OracleID:     "ab42398c-f0a1-4b94-ac5f-b8768e1b4e05",
		Name:         "Presence of Gond",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Activated: []effects.ActivatedAbility{{
				Label:   text,
				Cost:    effects.TapCost(),
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerValue},
				Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green Elf Warrior"), N: 1}),
			}},
			Text: text,
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(grant)},
	})
}
