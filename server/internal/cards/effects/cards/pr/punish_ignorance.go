package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Punish Ignorance — Instant {W}{U}{U}{B} (Shards of Alara, S82 OLD batch 518,
// #5588):
//
//	"Counter target spell. Its controller loses 3 life and you gain 3 life."
//
// Undermine's sting with Absorb's gain: the spell's controller is read before
// the counter moves it off the stack, the loss is life loss and not damage,
// and the gain is the caster's. A target that left the stack in response makes
// the whole spell fizzle (CR 608.2b), so neither life change happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0fc56124-bdcd-4f4f-886f-0fafcf8003bc",
		Name:         "Punish Ignorance",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, LifeGain: 3},
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				victim := uuid.Nil
				if si := ctx.Game.StackItemForEffect(t.ID); si != nil {
					victim = si.Controller
				}
				if err := (effects.CounterTarget{StackID: t.ID}).Apply(ctx); err != nil {
					return err
				}
				if victim != uuid.Nil {
					if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), victim, -3); err != nil {
						return err
					}
				}
				return effects.GainLife{Amount: 3}.Apply(ctx)
			}
			return nil
		},
	})
}
