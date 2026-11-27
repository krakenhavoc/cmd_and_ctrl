package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fortitude — Enchantment — Aura {1}{G} (Urza's Saga, #5624):
//
//	"Enchant creature
//	 Sacrifice a Forest: Regenerate enchanted creature.
//	 When this Aura is put into a graveyard from the battlefield, return it to
//	 its owner's hand."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5017c07d-8ad5-4ee2-91d3-974811cdff75",
		Name:         "Fortitude",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice a Forest: Regenerate enchanted creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.SacrificeN(1, "a Forest", effects.Subtype("Forest")),
			Effect:  effects.RegenerateAttachedHost,
		}},
		Triggered: []game.TriggeredAbility{effects.UlgB1AuraReturnsToHand("Fortitude")},
	})
}
