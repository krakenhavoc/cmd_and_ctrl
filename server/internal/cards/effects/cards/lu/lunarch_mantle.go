package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const lunarchMantleGrant = "lunarch-mantle/flying"

// Lunarch Mantle — Enchantment — Aura {1}{W} (Innistrad Remastered):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2 and has "{1}, Sacrifice a permanent: This
//	 creature gains flying until end of turn.""
//
// The granted ability is the creature's, so "this creature" is the host and
// its controller pays, whoever controls the Aura.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0da104ae-c46c-4106-919f-d975721d42d9",
		Name:         "Lunarch Mantle",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: lunarchMantleGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{1}, Sacrifice a permanent: This creature gains flying until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant | game.AnswerSacOutlet},
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeAPermanent()),
				Effect:  effects.ThisGainsKeywordUntilEndOfTurn("flying", "Lunarch Mantle — flying until end of turn"),
			}},
			Text: "{1}, Sacrifice a permanent: This creature gains flying until end of turn.",
		}},
		Static: []game.StaticAbility{effects.PumpAttached(2, 2), effects.GrantAbilitiesToAttached(lunarchMantleGrant)},
	})
}
