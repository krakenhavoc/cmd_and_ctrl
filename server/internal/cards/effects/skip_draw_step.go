package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// skip_draw_step.go — "Skip your draw step" (CR 614.10 /
// CR 500.11), the price half of Necropotence and Yawgmoth's Bargain.
// Its own file rather than helpers.go, per the convention #231 set:
// concurrent card batches collide on shared helper files.

// SkipYourDrawStep is the replacement effect behind "Skip your draw
// step." It cancels the step-transition event for the CONTROLLER's
// draw step and leaves every other seat's alone.
//
// Three things this is not, and each is a bug someone would
// otherwise write:
//
//   - Not "skip your draw". CR 500.11 skips the whole STEP, so the
//     draw step grants no priority, nothing can be cast in it, and
//     "at the beginning of each player's draw step" triggers
//     (Howling Mine) do not fire for that player. The engine's
//     step-transition cancel path in runStepEntryHooksLocked does
//     exactly that: it advances the cursor past the step and
//     recurses, so the turn walks upkeep → precombat main.
//
//   - Not symmetric. Stasis's untap skip fires on any seat's step
//     because its text says "players"; Necropotence says "your", so
//     the seat being entered has to be compared against the source's
//     controller. Omitting that check locks the whole table's draws
//     off a single enchantment.
//
//   - Not conditional on the turn-1 skip. In a TWO-player game the
//     engine's StepDraw hook already skips the starting player's
//     first draw step (CR 103.8a; at three or more seats CR 103.8c
//     has nobody skip), and two skips of the same step are one skip
//     — a replacement that fires on an already-skipped step changes
//     nothing (CR 614.5).
//
// Declared on Spec.Replacements, like every other continuous
// replacement a permanent contributes: the effect exists for as long
// as the permanent is on the battlefield (CR 113.6), so destroying
// the enchantment restores the draw from the next turn on.
func SkipYourDrawStep() game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventStepTransition},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventStepTransition || ev.StepTransitionStep != game.StepDraw {
				return false
			}
			return src != nil && seatIsPlayer(g, ev.StepTransitionSeat, src.Controller)
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.Cancel()
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: "Skip your draw step",
	}
}

// seatIsPlayer reports whether seat index `seat` is held by
// `playerID`. Out-of-range seats are false rather than a panic: the
// index arrives on a replacement event built from the turn cursor,
// and a cursor pointing nowhere (pre-game, an emptied table) is a
// state this predicate has to survive rather than diagnose.
func seatIsPlayer(g *game.Game, seat int, playerID uuid.UUID) bool {
	if g == nil || seat < 0 || seat >= len(g.Seats) {
		return false
	}
	p := g.Seats[seat]
	return p != nil && p.ID == playerID
}
