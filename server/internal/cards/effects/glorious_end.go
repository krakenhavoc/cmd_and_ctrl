package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glorious End — {2}{R} Instant:
//
//	"End the turn. (Exile all spells and abilities from the stack,
//	 including this card. The player whose turn it is discards down to
//	 their maximum hand size. Damage wears off, and "this turn" and
//	 "until end of turn" effects end.)
//	 At the beginning of your next end step, you lose the game."
//
// Time Stop at half the price, paid for later. EndTheTurn
// (end_the_turn.go, CR 724.1, #2165), and a CR 603.7 delayed trigger
// for the second sentence:
//
//   - "YOUR next end step" is ControllerTurnOnly: an opponent's end step
//     does not fire it.
//   - The end step of the turn Glorious End ends never begins
//     (CR 724.1e), so even cast on your own turn, before your end step,
//     the trigger waits for your next turn's. Cast during your own end
//     step, that step has already begun, and the drain that would fire
//     it ran as it began; it waits for the next one either way.
//   - It is a triggered ability, so it uses the stack and can be
//     answered; the loss goes through ADR 0057's LoseTheGame, so a
//     "you can't lose the game" effect stops it. Ending that turn too
//     (a second Glorious End, Sundial of the Infinite) skips that end
//     step and the trigger waits again.
//
// The delayed trigger is scheduled BEFORE the turn ends, against the
// printed order, because ending the turn exiles this spell and must be
// the last thing it does. Nothing about the turn ending reads or
// changes the delayed trigger queue, so the two orders are the same
// game. The body is the one Final Fortune's loss already registers
// (its controller loses the game); only its key's area says
// "extra-turn".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e611a3e0-eb0e-466e-a771-51310eeb34cd",
		Name:         "Glorious End",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (ScheduleDelayedTrigger{
				At:                 game.StepEnd,
				ControllerTurnOnly: true,
				Label:              "Glorious End — you lose the game",
				Body:               loseTheGameAtThatTurnsEndBody,
			}).Apply(ctx); err != nil {
				return err
			}
			return EndTheTurn{}.Apply(ctx)
		},
	})
}
