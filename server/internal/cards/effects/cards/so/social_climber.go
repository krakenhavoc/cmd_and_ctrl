package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Social Climber — Creature — Human Druid {2}{G}, 3/2 (Streets of New
// Capenna, #5491):
//
//	"Alliance — Whenever another creature you control enters, you gain 1
//	 life."
//
// Distinguished Conjurer's trigger: another creature you control (tokens
// count, the Climber's own entry and an opponent's creatures do not).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "70af9a03-20d6-44e3-a181-b60e80bff643",
		Name:         "Social Climber",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AnotherCreatureEnteredUnderYourControl,
				"Social Climber — alliance: you gain 1 life",
				effects.Do(effects.GainLife{Amount: 1})),
		},
	})
}
