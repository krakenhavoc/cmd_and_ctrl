package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// SkipOpponentsExtraTurns is "if an opponent would begin an extra turn,
// that player skips that turn instead" (CR 614.10; Trouble in Pairs,
// #2529). Declare it on Spec.Replacements.
//
// It replaces a RepEventExtraTurn: the window popExtraTurnLocked opens
// over a queued extra turn at the instant it would begin. Cancelling it
// is the skip. The turn never begins, so nothing bound to it fires (CR
// 614.10a — an opponent's Final Fortune never loses them the game), and
// the next queued turn, or normal rotation, follows.
//
// "Opponent" is any player other than the permanent's controller, which
// is how Kismet reads it. The controller's OWN extra turns are not
// touched: the card says "an opponent", and Ugin's Nexus, the card that
// says "a player", is the variant that would drop that comparison.
//
// PureCancel: Replace does nothing but cancel, so two of these on one
// table (two Troubles in Pairs, under different controllers) have no
// CR 616 ordering to ask about — the turn is skipped whichever applies
// first. The window cannot ask in any case (it sets mustSettleNow).
func SkipOpponentsExtraTurns() game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventExtraTurnBegin},
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventExtraTurn || src == nil {
				return false
			}
			return ev.Actor != uuid.Nil && ev.Actor != src.Controller
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.Cancel()
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		PureCancel: true,
		Label:      "Skip an opponent's extra turn",
	}
}
