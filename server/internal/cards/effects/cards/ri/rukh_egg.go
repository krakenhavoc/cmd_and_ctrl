package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rukh Egg — Creature — Bird Egg {3}{R} 0/3 (Eighth Edition, S82 older sets batch 620, #5607):
//
//	"When this creature dies, create a 4/4 red Bird creature token with flying
//	 at the beginning of the next end step."
//
// Roc Egg's dies trigger with the token made later: the death schedules a
// delayed trigger (CR 603.7) for the next end step, controlled by whoever
// controlled the Egg, and the token is made then even though the Egg is long
// gone. The token is the Oracle's Bird; the only printed token card is the
// older Rukh (see knownUnmatchedTokens).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "98116aec-2ab1-4bee-b727-9feff6274825",
		Name:         "Rukh Egg",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Rukh Egg — create a 4/4 red Bird creature token with flying at the beginning of the next end step",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ScheduleDelayedTrigger{
						At:    game.StepEnd,
						Label: "Rukh Egg — create a 4/4 red Bird creature token with flying",
						Body:  rukhEggBirdBody,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// rukhEggBirdBody is the delayed trigger's body. An on-disk key: never
// renamed, never reused.
var rukhEggBirdBody = game.SimpleDelayedBody("rukh-egg/create-bird", func(g *game.Game, item *game.StackItem) error {
	return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("4/4 red Bird with flying"), N: 1}.Apply(effects.NewContext(g, item))
})
