package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "0ce01e24-42db-42db-9ba2-38f653383991",
		Name:         "Gravebind",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := targetCantBeRegeneratedThisTurn(ctx.Game, item); err != nil {
				return err
			}
			return ScheduleDelayedTrigger{
				At:    game.StepUpkeep,
				Label: "Gravebind — draw a card",
				Body:  drawOneBody,
			}.Apply(ctx)
		},
	})
}
