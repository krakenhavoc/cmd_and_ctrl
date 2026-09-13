package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Evacuation — Instant {3}{U}{U}:
//
//	"Return all creatures to their owners' hands."
//
// The blue wrath, and the one that is an INSTANT — cast in response
// to the alpha strike, or at the end of the turn before yours. Five
// mana rather than four is the price for that plus the fact that
// nothing dies.
//
// What bounce buys over destruction, all of it relevant at a
// Commander table:
//
//   - Indestructible, regeneration and "can't be destroyed" are
//     irrelevant; the permanent leaves the battlefield regardless.
//   - Tokens cease to exist entirely rather than becoming graveyard
//     fuel.
//   - No dies-triggers, so a table full of aristocrats payoffs does
//     not drain anyone.
//   - And the cost: everyone REBUYS their board, so against a deck
//     of cheap creatures this is a tempo play, not an answer.
//
// Symmetric — your own creatures come back to your hand too.
func init() {
	Register(Spec{
		OracleID:     "fdd94383-b573-439a-8e1c-925af887c5a6",
		Name:         "Evacuation",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return BounceAllMatching{Match: Creature()}.Apply(ctx)
		},
	})
}
