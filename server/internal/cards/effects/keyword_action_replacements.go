package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// keyword_action_replacements.go — the CR 614 replacement on a KEYWORD
// ACTION with a count: "if you would proliferate, proliferate twice
// instead" (Tekuthal, Inquiry Dominus), "if you would scry, scry that
// many plus one instead" (the Crystal Ball family), and the same
// sentence for surveil. Its own file rather than helpers.go, per the
// convention enters_tapped.go set.
//
// The event is game.RepEventKeywordAction (#976), opened once per
// INSTRUCTION at the one entry point of each action — the way
// RepEventCreateTokens is opened once per creation instruction. It
// carries which action it is, the player taking it and the card whose
// effect asked, and ONE field a replacement rewrites: the count. What
// the count means is per action and is written on the
// game.KeywordAction constants — for proliferate it is the number of
// TIMES the action is taken, for scry and surveil the number of cards.
//
// So the printed family is one function with one argument that
// differs: what the count becomes.

// KeywordActionBecomes is "if you would <action>, <action> <count(n)>
// instead" — the whole family, with the arithmetic as a function.
//
// `count` is applied to the count the event currently carries, NOT to
// the printed one, and that is the rules' own composition: CR 616.1
// applies one replacement and then re-gathers, so a doubler and a
// "plus one" in the same window compose in the order the affected
// player picks (×2 then +1 is 3 proliferates from a base of 1; +1 then
// ×2 is 4). Two copies of the SAME card do not ask — they are two
// objects contributing one declared effect, and #792's identical-window
// skip applies — so two Tekuthals are four proliferates, no prompt.
//
// Scoped to the source's controller, because every printed one says
// "if YOU would". A symmetrical printing would need an AnyPlayer
// sibling here, the way AnyPlayersTokensDoubled is TokensDoubled's;
// none exists yet.
func KeywordActionBecomes(action game.KeywordAction, count func(n int) int, label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventKeywordAction},
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventKeywordAction || ev.KeywordAction != action {
				return false
			}
			if ev.KeywordActionCount <= 0 {
				// Already replaced down to nothing, or never asked
				// for: there is no action left to replace.
				return false
			}
			return ev.Actor == src.Controller
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			if count == nil {
				return nil
			}
			ev.KeywordActionCount = count(ev.KeywordActionCount)
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: label,
	}
}

// ProliferateTwice is "if you would proliferate, proliferate twice
// instead" — Tekuthal, Inquiry Dominus. The count is the number of
// TIMES the keyword action is taken (CR 701.34), so this doubles the
// whole action, choice and all, rather than the counters one
// proliferate places.
func ProliferateTwice(label string) game.ReplacementEffect {
	return KeywordActionBecomes(game.KeywordActionProliferate, timesTwo, label)
}

// ScryPlusOne is "if you would scry, scry that many plus one instead"
// (CR 701.22) — the Crystal Ball family. "That many" is the printed
// count, which is why this adds to the event's count rather than
// setting it.
func ScryPlusOne(label string) game.ReplacementEffect {
	return KeywordActionBecomes(game.KeywordActionScry, plusOne, label)
}

// SurveilPlusOne is ScryPlusOne's other half — "if you would surveil,
// surveil that many plus one instead" (CR 701.25).
func SurveilPlusOne(label string) game.ReplacementEffect {
	return KeywordActionBecomes(game.KeywordActionSurveil, plusOne, label)
}

// timesTwo and plusOne are the two arithmetics the printed cards use.
// Named rather than written inline at each call so the three helpers
// above are one line each and a fourth card reaches for one of these
// before inventing a third.
func timesTwo(n int) int { return n * 2 }

func plusOne(n int) int { return n + 1 }
