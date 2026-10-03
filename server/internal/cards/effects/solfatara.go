package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Solfatara — Instant {2}{R}:
//
//	"Target player can't play lands this turn.
//	 Draw a card at the beginning of the next turn's upkeep."
//
// ADR 0109 §4 (#1895). Turf Wound's ban (CantPlayLandsThisTurn), then a
// CR 603.7 delayed trigger for the very next upkeep at the table,
// whoever's it is ("the next turn's upkeep", Urza's Bauble's wording and
// code). The draw goes on the stack when that upkeep begins and is the
// caster's (CR 603.7d). The delayed trigger is scheduled even if the
// target became illegal, because only an all-illegal spell fizzles.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6053a192-fcf9-4b06-9f46-e84eaadaa882",
		Name:         "Solfatara",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (CantPlayLandsThisTurn{}).Apply(ctx); err != nil {
				return err
			}
			return ScheduleDelayedTrigger{
				At:    game.StepUpkeep,
				Label: "Solfatara — draw a card",
				Body:  drawOneBody,
			}.Apply(ctx)
		},
	})
}
