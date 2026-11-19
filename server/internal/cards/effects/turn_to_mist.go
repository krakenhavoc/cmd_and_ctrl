package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "631f04d6-5693-4414-a378-fd839c3b6f6d",
		Name:         "Turn to Mist",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return exileCardsThenScheduleReturn(ctx, legalTargetIDs(ctx), "Turn to Mist — return the exiled creature")
		},
	})
}
