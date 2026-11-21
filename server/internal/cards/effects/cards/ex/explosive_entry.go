package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Explosive Entry — Sorcery {1}{R} (Kamigawa: Neon Dynasty, #5493):
//
//	"Destroy up to one target artifact. Put a +1/+1 counter on up to one
//	 target creature."
//
// Two independent "up to one" clauses: the spell can be cast with either,
// both, or neither. Each is re-checked as it resolves (CR 608.2b), so one
// going illegal leaves the other to resolve. The artifact can be the
// creature's own (an artifact creature) and both halves still happen.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "57544b37-fd5d-4da5-bbb6-7b4a09501dd1",
		Name:         "Explosive Entry",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: effects.TargetPermanent("up to one target artifact", effects.Artifact()).WithCount(0, 1).
			Then(effects.TargetCreature("up to one target creature").WithCount(0, 1)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			// The counter goes first so the destroy is the last act and
			// nothing is gated on its outcome; the two halves are
			// independent, so the order is not observable.
			if t, ok := ctx.ClauseTarget(1); ok {
				if err := (effects.AddCounter{Target: t.ID, Kind: "+1/+1", N: 1}).Apply(ctx); err != nil {
					return err
				}
			}
			if t, ok := ctx.ClauseTarget(0); ok {
				return effects.DestroyTarget{Target: t.ID}.Apply(ctx)
			}
			return nil
		},
	})
}
