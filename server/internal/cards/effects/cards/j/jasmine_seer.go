package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jasmine Seer — Creature — Human Wizard {3}{W}, 1/1 (Urza's Destiny, S82 #5621):
//
//	"{2}{W}, {T}: Reveal any number of white cards in your hand. You gain 2 life
//	 for each card revealed this way."
//
// Scent of Jasmine's body as an activated ability (OldB706RevealAnyNumber): the
// reveal is chosen as the ability resolves, and the life is gained even if the
// Seer has left.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d934a7f0-eb46-4a2c-b0fd-e3071e23da4a",
		Name:         "Jasmine Seer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}, {T}: Reveal any number of white cards in your hand. You gain 2 life for each card revealed this way.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}{W}"), effects.TapCost()),
			Effect:  effects.OldB706AsEffect(effects.OldB706RevealAnyNumber("Jasmine Seer", "W", "choose-cards/s82-old-b706-jasmine-seer", effects.OldB706GainTwoPerRevealed)),
		}},
	})
}
