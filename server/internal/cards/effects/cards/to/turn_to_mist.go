package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Turn to Mist — Instant {1}{W/U} (Ravnica: Clue Edition):
//
//	"Exile target creature. Return that card to the battlefield under its
//	 owner's control at the beginning of the next end step."
//
// Voidwalk's delayed blink (a CR 603.7 delayed trigger, the card returning
// under its owner's control), on any creature rather than one the caster
// controls. A token exiled this way ceases to exist and does not return.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "631f04d6-5693-4414-a378-fd839c3b6f6d",
		Name:         "Turn to Mist",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ExileCardsThenScheduleReturn(ctx, effects.LegalTargetIDs(ctx), "Turn to Mist — return the exiled creature")
		},
	})
}
