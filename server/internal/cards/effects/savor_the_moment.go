package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Savor the Moment — Sorcery {1}{U}{U}:
//
//	"Take an extra turn after this one. Skip the untap step of that turn."
//
// TakeExtraTurn with SkipSteps (ADR 0188 Decision 3a, #3563, tracker
// #753): the turn is queued (CR 500.7) and a skipExtraTurnStep record is
// bound to it. The record is spent at the step's entry, before the
// CR 614 window opens, and the cursor walks past the step as it does a
// cancelled one (CR 500.11), so in the extra turn nothing untaps,
// nothing phases in (CR 502.1) and the day/night check does not run
// (CR 502.2). What it does NOT touch:
//
//   - Summoning sickness is cleared as the turn begins, not in the untap
//     step, so last turn's creatures can attack.
//   - A "doesn't untap during its controller's next untap step" marker
//     waits for the next untap step that really happens (CR 614.10a).
//   - The next turn's day/night check reads this turn (CR 731.2): an
//     extra turn is a turn, and its spells count.
//   - Only the first untap step is skipped. A beginning phase added to
//     that turn (Sphinx of the Second Sun) has an untap step that
//     happens (owner answer 2 of ADR 0188).
//   - If the extra turn is itself skipped (Trouble in Pairs, Ugin's
//     Nexus, a skip debt) the record goes with it (CR 614.10a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d5b9b75d-4bf0-4328-9bb3-7b260a0c46af",
		Name:         "Savor the Moment",
		Completeness: CompletenessFull,
		// ADR 0188 Decision 10: the caster's extra turn, priced by the bot.
		Purpose: game.Purpose{ExtraTurns: 1},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return TakeExtraTurn{
				SkipSteps: []game.Step{game.StepUntap},
				Label:     "Savor the Moment — skip the untap step of that turn",
			}.Apply(ctx)
		},
	})
}
