package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// token_replacements.go — the CR 701.7b replacement on "create one or
// more tokens". Its own file rather than helpers.go, per the
// convention enters_tapped.go set: concurrent card batches collide on
// shared helper files, and the token doubler is a keystone several of
// them want.
//
// The event is game.RepEventCreateTokens, opened once per creation
// INSTRUCTION (#762, ADR 0061). A doubler multiplies its counts; a
// card that changes WHAT is made rewrites its kind set. Both are one
// line on the event — see game.ReplacementEvent.MultiplyTokens and
// ReplaceTokenKinds — so the card files below are the printed
// sentence and nothing else.

// TokensDoubled is "if an effect would create one or more tokens
// under your control, it creates twice that many of those tokens
// instead" — Doubling Season, Parallel Lives, Anointed Procession,
// Mondrak. `n` is the multiplier, which is 2 on every printed card
// that has one.
//
// Multiplicative by construction, which is what the rules make two of
// them: CR 616.1 applies one and then the other to the same event, so
// two Anointed Processions are ×4 whichever is named first. The
// affected player is not asked to order two copies of one declared
// effect (#792's identical-window skip), and IS asked when a
// different effect shares the window — an Academy Manufactor, say —
// because those orderings really do differ.
func TokensDoubled(label string) game.ReplacementEffect {
	return tokenMultiplier(label, 2, true)
}

// AnyPlayersTokensDoubled is TokensDoubled without the controller
// clause — "if one or more tokens would be created, twice that many
// of those tokens are created instead" (Primal Vigor). Symmetrical:
// every player's tokens, the source's controller included.
func AnyPlayersTokensDoubled(label string) game.ReplacementEffect {
	return tokenMultiplier(label, 2, false)
}

// tokenMultiplier is the shared body. `yoursOnly` narrows the event to
// creations under the source's controller.
func tokenMultiplier(label string, n int, yoursOnly bool) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventTokenCreated},
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventCreateTokens || ev.TokenCount() <= 0 {
				return false
			}
			return !yoursOnly || ev.TokenController == src.Controller
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.MultiplyTokens(n)
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: label,
	}
}
