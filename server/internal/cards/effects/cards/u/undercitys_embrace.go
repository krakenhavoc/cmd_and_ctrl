package cardsu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Undercity's Embrace — Instant {2}{B} (Ravnica Remastered):
//
//	"Target opponent sacrifices a creature of their choice. If you control
//	 a creature with power 4 or greater, you gain 4 life."
//
// The sacrifice is the opponent's prompt and goes first; the life clause
// reads my creatures as it resolves, whether or not the opponent had a
// creature to sacrifice.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5b7b1839-2b44-4446-80b4-ddb59e4e835e",
		Name:         "Undercity's Embrace",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				who, ok := ctx.ClauseTarget(0)
				if !ok || who.Kind != game.TargetPlayer {
					return nil
				}
				ctx.Game.PlayersSacrificeForEffect(ctx.Source(), []uuid.UUID{who.ID},
					effects.SacrificeSpec("a creature", effects.Creature()),
					"Undercity's Embrace — sacrifice a creature")
				return nil
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				if effects.CountControlled(ctx.Game, ctx.Controller(), func(c game.Card) bool {
					return c.IsCreature() && c.CurrentPower() >= 4
				}) == 0 {
					return nil
				}
				return effects.GainLife{Amount: 4}.Apply(ctx)
			}),
		),
	})
}
