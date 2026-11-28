package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Telim'Tor's Edict — Instant {R} (Mirage, S82 OLD b800, #5633):
//
//	"Exile target permanent you own or control.
//	 Draw a card at the beginning of the next turn's upkeep."
//
// The target is any permanent the caster owns or controls (a stolen permanent
// they own is as legal as one they control); it is exiled as the spell
// resolves. The draw is Burnout's delayed trigger at the next upkeep (CR
// 603.7a), scheduled whenever the spell resolves; a spell whose only target is
// gone does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0e768d03-b2a3-4783-9cf5-7470f0203349",
		Name:         "Telim'Tor's Edict",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target permanent you own or control", effects.Or(effects.YouControl(), effects.YouOwn())),
		Purpose:      game.Purpose{Answers: game.AnswerProtect, Draws: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.ExileFirstLegalCardTarget(ctx.Game, item); err != nil {
				return err
			}
			return effects.DrawACardAtTheNextUpkeep("Telim'Tor's Edict — draw a card").Apply(ctx)
		},
	})
}
