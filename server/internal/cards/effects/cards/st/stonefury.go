package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stonefury — Instant {3}{R}{R} (Battle for Zendikar, #5547):
//
//	"Stonefury deals damage to target creature equal to the number
//	 of lands you control."
//
// Rumbling Rockslide's body at instant speed: the lands are counted as
// it resolves (CR 608.2h), and a target that has left by then is not
// dealt damage (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "577554de-78c3-4fa2-a866-06b80280bc97",
		Name:         "Stonefury",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
			lands := effects.CountControlled(ctx.Game, ctx.Controller(), func(c game.Card) bool { return c.IsLand() })
			return effects.DealDamage{Source: ctx.Source(), Target: id, Amount: lands}.Apply(ctx)
		})),
	})
}
