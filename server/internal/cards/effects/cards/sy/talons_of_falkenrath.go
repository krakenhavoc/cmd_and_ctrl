package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Talons of Falkenrath — Enchantment — Aura {1}{R} (Dark Ascension, S82 OLD,
// #5571):
//
//	"Flash
//	 Enchant creature
//	 Enchanted creature has "{1}{R}: This creature gets +2/+0 until end of
//	 turn.""
//
// Presence of Gond's granted activated ability: the creature, not the Aura,
// is the source, so its controller pays and activates it, and the grant goes
// with the Aura. Flash is the printed keyword.
//
// No simplification.
func init() {
	const grant = "talons-of-falkenrath/pump"
	const text = "{1}{R}: This creature gets +2/+0 until end of turn."
	effects.Register(effects.Spec{
		OracleID:        "e79b1224-1baf-460f-bf12-92b51cd8071f",
		Name:            "Talons of Falkenrath",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Activated: []effects.ActivatedAbility{{
				Label:   text,
				Cost:    effects.ManaCost("{1}{R}"),
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Effect:  effects.ThisGetsUntilEndOfTurn(2, 0, "Talons of Falkenrath — +2/+0"),
			}},
			Text: text,
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(grant)},
	})
}
