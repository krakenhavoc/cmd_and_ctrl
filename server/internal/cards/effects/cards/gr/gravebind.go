package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gravebind — Instant {B}:
//
//	"Target creature can't be regenerated this turn.
//	 Draw a card at the beginning of the next turn's upkeep."
//
// The draw is a CR 603.7 delayed trigger at the next upkeep at the table,
// whoever's it is (Portent's reading), controlled by Gravebind's
// controller. It is scheduled as the spell resolves.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0ce01e24-42db-42db-9ba2-38f653383991",
		Name:         "Gravebind",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.TargetCantBeRegeneratedThisTurn(ctx.Game, item); err != nil {
				return err
			}
			return effects.ScheduleDelayedTrigger{
				At:    game.StepUpkeep,
				Label: "Gravebind — draw a card",
				Body:  effects.DrawOneBody,
			}.Apply(ctx)
		},
	})
}
