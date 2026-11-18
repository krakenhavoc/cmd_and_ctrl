package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "70af9a03-20d6-44e3-a181-b60e80bff643",
		Name:         "Social Climber",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AnotherCreatureEnteredUnderYourControl,
				"Social Climber — alliance: you gain 1 life",
				Do(GainLife{Amount: 1})),
		},
	})
}
