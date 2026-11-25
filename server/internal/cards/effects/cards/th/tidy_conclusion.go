package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tidy Conclusion — Instant {3}{B}{B} (Kaladesh, #5539):
//
//	"Destroy target creature. You gain 1 life for each artifact you control."
//
// Venom's Hunger's destroy-then-life shape: the artifacts are counted after
// the destroy, so an artifact creature that dies to it is not counted, and the
// life is gained whether or not the creature was destroyed (indestructible);
// an illegal target fizzles the whole spell (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d73adc2e-c458-40a9-b10b-7d2133778717",
		Name:         "Tidy Conclusion",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			if err := (effects.DestroyTarget{Target: item.Targets[0].ID}).Apply(ctx); err != nil {
				return err
			}
			n := len(effects.PermanentsControlledByMatching(ctx.Game, ctx.Controller(), effects.Artifact()))
			return effects.GainLife{Player: ctx.Controller(), Amount: n}.Apply(ctx)
		},
	})
}
