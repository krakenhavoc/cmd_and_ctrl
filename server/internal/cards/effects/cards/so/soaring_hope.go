package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soaring Hope — Enchantment — Aura {4}{W} (Lorwyn, S82 OLD, #5593):
//
//	"Enchant creature
//	 When this Aura enters, you gain 3 life.
//	 Enchanted creature has flying.
//	 {W}: Put this Aura on top of its owner's library."
//
// Angelic Gift's flying grant, an enters trigger that gains the life, and
// Wayward Soul's tuck as the Aura's own activated ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0d09049a-6786-4b69-b448-35efaa926a71",
		Name:         "Soaring Hope",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.GrantToAttached("flying")},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Soaring Hope — you gain 3 life", effects.Do(effects.GainLife{Amount: 3})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}: Put this Aura on top of its owner's library.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{W}"),
			Effect:  effects.PutThisPermanentOnTopOfOwnersLibrary,
		}},
	})
}
