package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Self-Inflicted Wound — Sorcery {1}{B} (Dragons of Tarkir, #5551):
//
//	"Target opponent sacrifices a green or white creature of their choice. If that player does, they lose 2 life."
//
// The opponent picks among their green and white creatures (CR 701.21a); the
// life is lost only when a creature was really sacrificed.
//
// No simplification.

func init() {
	effects.Register(effects.Spec{
		OracleID:     "28956db0-985c-402d-abae-1075c68b97e7",
		Name:         "Self-Inflicted Wound",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player, ok := effects.FirstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			return ctx.PlayersSacrificeThen([]uuid.UUID{player},
				effects.SacrificeSpec("a green or white creature", effects.And(effects.Creature(), effects.Or(effects.OfColor("G"), effects.OfColor("W")))),
				"Self-Inflicted Wound — sacrifice a green or white creature",
				selfInflictedWoundLose, ctx.RunData([]uuid.UUID{player}))
		},
	})
}

// selfInflictedWoundLose is "if that player does, they lose 2 life": the
// carry names the player, and the run says whether they sacrificed.
var selfInflictedWoundLose = effects.RunThen("prompt-run/self-inflicted-wound-lose", func(ctx *effects.Context, r game.RunDone) error {
	if len(r.Carry) == 0 || !r.Sacrifices().Sacrificed(r.Carry[0]) {
		return nil
	}
	return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), r.Carry[0], -2)
})
