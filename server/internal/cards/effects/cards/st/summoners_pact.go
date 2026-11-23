package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Summoner's Pact — Instant {0}:
//
//	"Search your library for a green creature card, reveal it, put it into
//	 your hand, then shuffle.
//	 At the beginning of your next upkeep, pay {2}{G}{G}. If you don't, you
//	 lose the game."
//
// Pact of Negation's debt (pactPaymentBody, a delayed trigger on your next
// upkeep) around Eladamri's Call's search. The debt is scheduled before the
// search so a library with no green creature still owes it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "782dcbc4-5a11-42f5-b8d9-4572e36d45b3",
		Name:         "Summoner's Pact",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := (effects.ScheduleDelayedTrigger{
				At:                 game.StepUpkeep,
				ControllerTurnOnly: true,
				Label:              "Summoner's Pact — pay {2}{G}{G} or lose the game",
				Body:               effects.PactPaymentBody,
				Params:             game.EffectParams{Cost: "{2}{G}{G}", Name: "Summoner's Pact"},
			}).Apply(ctx); err != nil {
				return err
			}
			return effects.B06TutorToHand("Summoner's Pact — a green creature card", func(c game.Card) bool {
				return c.IsCreature() && c.HasColorWord("G")
			})(item, ctx)
		},
	})
}
