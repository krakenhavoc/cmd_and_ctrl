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
// touched: the card says "an opponent". SkipExtraTurns is the variant
// that drops that comparison (Ugin's Nexus, Gerrard's Hourglass Pendant).
//
// PureCancel: Replace does nothing but cancel, so two of these on one
// table (two Troubles in Pairs, under different controllers) have no
// CR 616 ordering to ask about — the turn is skipped whichever applies
// first. The window cannot ask in any case (it sets mustSettleNow).
func SkipOpponentsExtraTurns() game.ReplacementEffect {
	return skipExtraTurnsReplacement("Skip an opponent's extra turn", func(ev *game.ReplacementEvent, src *game.Card) bool {
		return ev.Actor != src.Controller
	})
}

// SkipExtraTurns is "if a player would begin an extra turn, that player
// skips that turn instead" (CR 614.10; Ugin's Nexus, Gerrard's Hourglass
// Pendant; ADR 0188 Decision 2, #3560). Declare it on Spec.Replacements.
//
// It is SkipOpponentsExtraTurns without the opponent test: the card says
// "a player", so its controller's own extra turns are skipped too. Both
// build from one constructor, so the window, the pure cancel and the
// CR 616 behaviour cannot drift apart.
//
// The skip happens when the turn WOULD BEGIN, not when it is queued, so
// an extra turn queued while the permanent is on the battlefield is
// taken if the permanent has left by then (the Ugin's Nexus ruling of
// 2014-09-20), and one queued before the permanent arrived is skipped.
func SkipExtraTurns() game.ReplacementEffect {
	return skipExtraTurnsReplacement("Skip every extra turn", func(*game.ReplacementEvent, *game.Card) bool {
		return true
	})
}

// skipExtraTurnsReplacement is the one constructor behind both skips: a
// pure cancel of a RepEventExtraTurn whose taker passes `whose`. The
// taker must be a real player (ev.Actor non-zero).
func skipExtraTurnsReplacement(label string, whose func(ev *game.ReplacementEvent, src *game.Card) bool) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventExtraTurnBegin},
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventExtraTurn || src == nil || ev.Actor == uuid.Nil {
				return false
			}
			return whose(ev, src)
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.Cancel()
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		PureCancel: true,
		Label:      label,
	}
}
