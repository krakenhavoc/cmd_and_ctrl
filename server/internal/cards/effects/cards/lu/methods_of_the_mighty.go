package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Methods of the Mighty — Instant {3}{W} (Marvel Super Heroes
// Commander, #4455):
//
//	"Choose one or more —
//	 • Destroy target artifact.
//	 • Destroy target tapped creature.
//	 • Put a +1/+1 counter on each creature you control."
//
// Sublime Epiphany's "one or more" shape: bullets resolve in printed
// order, each with its own target group, and a bullet whose target
// became illegal is skipped without stopping the rest. The counters go
// on the creatures you control as the third bullet resolves, so a
// creature destroyed by an earlier bullet gets none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "45f91d6a-eaab-4179-ac1f-5dc9e647a3da",
		Name:         "Methods of the Mighty",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOneOrMore(
			effects.ModeDoing("Destroy target artifact.",
				effects.TargetPermanent("target artifact", effects.Artifact()),
				effects.DestroyTheModesTarget),
			effects.ModeDoing("Destroy target tapped creature.",
				effects.TargetCreature("target tapped creature", effects.B751Tapped()),
				effects.DestroyTheModesTarget),
			effects.ModeDoing("Put a +1/+1 counter on each creature you control.", nil,
				func(item *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.B11PutCounterOnEachCreatureYouControl(ctx.Game, item)
				}),
		),
	})
}
