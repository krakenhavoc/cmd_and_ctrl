package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ground Assault — Sorcery {R}{G} (Double Masters 2022, #5948):
//
//	"Ground Assault deals damage to target creature equal to the number of
//	 lands you control."
//
// The lands are counted as the spell resolves, so a land creature counts and
// an opponent's land does not. The damage is dealt by the spell. A target
// that is gone or no longer a creature makes the spell do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "13613f6e-a6e4-4493-b029-8ec11610c29f",
		Name:         "Ground Assault",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
			n := effects.CountControlled(ctx.Game, ctx.Controller(), game.Card.IsLand)
			return effects.DealDamage{Source: ctx.Source(), Target: id, Amount: n}.Apply(ctx)
		})),
	})
}
