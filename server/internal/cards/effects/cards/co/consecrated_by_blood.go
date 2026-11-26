package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Consecrated by Blood — Enchantment — Aura {2}{B}{B} (Magic Origins):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2 and has flying and "Sacrifice two other
//	 creatures: Regenerate this creature.""
//
// The pump and flying are the Aura's statics; the regeneration is a granted
// activated ability (Presence of Gond's shape), so the enchanted creature is
// the source: its controller pays the sacrifice of two creatures other than
// it, and the shield goes on it.
//
// No simplification.
func init() {
	const grant = "consecrated-by-blood/regenerate"
	const text = "Sacrifice two other creatures: Regenerate this creature."
	effects.Register(effects.Spec{
		OracleID:     "270508aa-b2e7-4066-be2f-4308cffe2b94",
		Name:         "Consecrated by Blood",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Activated: []effects.ActivatedAbility{{
				Label:   text,
				Cost:    effects.SacrificeAnotherN(2, "two other creatures", effects.Creature()),
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Effect:  effects.RegenerateThisSource,
			}},
			Text: text,
		}},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GrantToAttached("flying"),
			effects.GrantAbilitiesToAttached(grant),
		},
	})
}
