package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Healing Leaves — Instant {G} (Planar Chaos, #5595):
//
//	"Choose one —
//	 • Target player gains 3 life.
//	 • Prevent the next 3 damage that would be dealt to any target this turn."
//
// The first bullet gains the target player 3 life; the second is Goblin
// Legionnaire's prevention shield on the bullet's own target, used up by
// the next 3 damage that would be dealt to it this turn (CR 615.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9c675e7a-1d68-44b2-9c1a-543b73782437",
		Name:         "Healing Leaves",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Target player gains 3 life.", effects.TargetPlayer("target player"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					return effects.GainLife{Player: t.ID, Amount: 3}.Apply(ctx)
				}),
			effects.ModeWithPurpose(effects.ModeDoing("Prevent the next 3 damage that would be dealt to any target this turn.", effects.TargetAny(),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return effects.PreventNextDamage{Target: t.ID, Amount: 3, Label: "Healing Leaves: prevent the next 3 damage"}.Apply(ctx)
				}), game.Purpose{Answers: game.AnswerPrevent}),
		),
	})
}
