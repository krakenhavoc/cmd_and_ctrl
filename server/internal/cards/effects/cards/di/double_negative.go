package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Double Negative — Instant {U}{U}{R} (Alara Reborn, #5586):
//
//	"Counter up to two target spells."
//
// Negate's counter for each of zero, one or two targets chosen as it is cast; a
// target that has left the stack, or is no longer legal, is skipped as it
// resolves (CR 608.2b) and the rest are still countered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9fe57d74-0394-4263-a1b1-da9773b485a6",
		Name:         "Double Negative",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetSpell("up to two target spells").WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.CounterTarget{StackID: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
