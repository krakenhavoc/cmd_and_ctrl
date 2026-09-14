package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Price of Progress — Instant {1}{R} (EDHREC rank 2814):
//
//	"Price of Progress deals damage to each player equal to twice the
//	 number of nonbasic lands that player controls."
//
// The burn deck's answer to a Commander manabase. Every seated
// player — the caster included — takes twice their nonbasic land
// count, counts taken before any damage is dealt. "Nonbasic" is the
// printed supertype: a Snow-Covered Forest is basic, a Dryad Arbor is
// not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e9da499c-fa43-4e94-8395-5c030ff39502",
		Name:         "Price of Progress",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b26DamageEachPlayerTwiceTheirNonbasicLands(ctx)
		},
	})
}
