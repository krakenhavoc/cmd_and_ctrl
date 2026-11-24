package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chance for Glory — Instant {1}{R}{W} (Secret Lair Drop):
//
//	"Creatures you control gain indestructible. Take an extra turn after this
//	 one. At the beginning of that turn's end step, you lose the game."
//
// The first sentence has no stated duration, so it lasts until end of turn
// (CR 611.2a) over the creatures its controller controls as it resolves
// (CR 611.2c, Warrior's Oath's neighbour in the catalog); the rest is Final
// Fortune's extra turn with the loss at that turn's end step
// (ExtraTurnThenLoseAtItsEndStep).
//
// No simplification.
func init() {
	extraTurn := effects.ExtraTurnThenLoseAtItsEndStep("Chance for Glory — you lose the game")
	effects.Register(effects.Spec{
		OracleID:     "324af60d-142b-4b6c-afe9-6aa36267d5ee",
		Name:         "Chance for Glory",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerSurvive},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := (effects.GrantKeywordUntilEOT{
				Match:    effects.And(effects.Creature(), effects.YouControl()),
				Keywords: []string{"indestructible"},
				Label:    "Chance for Glory — creatures you control gain indestructible until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			return extraTurn(item, ctx)
		},
	})
}
