package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eyes in the Skies — Instant {3}{W}:
//
//	"Create a 1/1 white Bird creature token with flying, then
//	 populate. (Create a token that's a copy of a creature token you
//	 control.)"
//
// The token is made first, so it is a candidate for the populate that
// follows (and the only one when you control no other creature token).
// Populate is the shared primitive in populate.go.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9697cd05-474e-46f3-8bcc-d4b6fb2059a1",
		Name:         "Eyes in the Skies",
		Completeness: effects.CompletenessFull,
		// Populate waits for a prompt the token's creation asks (a CR 616 order; CR 608.2c, #4155).
		OnResolve: effects.ResolveSteps(
			effects.CreateToken{Template: effects.TokenCard("1/1 white Bird with flying"), N: 1},
			effects.Populate{},
		),
	})
}
