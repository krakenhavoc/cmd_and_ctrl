package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Horncaller's Chant — Sorcery {7}{G}:
//
//	"Create a 4/4 green Rhino creature token with trample, then
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
		OracleID:     "8c069834-ed89-4298-989b-e67036d56196",
		Name:         "Horncaller's Chant",
		Completeness: effects.CompletenessFull,
		// Populate waits for a prompt the token asks (a CR 616 order; CR 608.2c, #4155).
		OnResolve: effects.ResolveSteps(
			effects.CreateToken{Template: effects.TokenCard("4/4 green Rhino with trample"), N: 1},
			effects.Populate{},
		),
	})
}
