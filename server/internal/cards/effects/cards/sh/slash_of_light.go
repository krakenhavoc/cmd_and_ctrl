package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Slash of Light — Instant {1}{W} (Final Fantasy, #4462):
//
//	"Slash of Light deals damage equal to the number of creatures you
//	 control plus the number of Equipment you control to target
//	 creature."
//
// Team Avatar's count with a second term, read as the spell resolves
// (CR 608.2h). The two counts are separate, as printed: an Equipment
// that is also a creature (a Reconfigure creature, an animated
// Equipment) adds to both. With its target gone the spell does nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3cc4d6f4-79b2-4210-84de-a0febf81eb67",
		Name:         "Slash of Light",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			target := effects.FirstLegalBattlefieldTarget(ctx)
			if target == uuid.Nil {
				return nil
			}
			creatures := effects.CountControlled(ctx.Game, item.Controller, func(c game.Card) bool { return c.IsCreature() })
			equipment := effects.CountControlled(ctx.Game, item.Controller, func(c game.Card) bool { return c.HasSubtype("Equipment") })
			return effects.DealDamage{Source: ctx.Source(), Target: target, Amount: creatures + equipment}.Apply(ctx)
		},
	})
}
