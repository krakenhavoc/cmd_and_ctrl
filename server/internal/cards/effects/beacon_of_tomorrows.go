package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Beacon of Tomorrows — Sorcery {6}{U}{U}:
//
//	"Target player takes an extra turn after this one. Shuffle Beacon
//	 of Tomorrows into its owner's library."
//
// Time Warp's first sentence (CR 500.7, targetPlayerTakesExtraTurns)
// and Blue Sun's Zenith's second: the resolving card tucks itself into
// its owner's library and shuffles. The card is still on the stack when
// OnResolve runs, so tucking it here is the "the spell moved itself"
// case resolveTopOfStackLocked already special-cases (CR 608.2n) and it
// skips the graveyard route.
//
// The extra turn is queued BEFORE the shuffle, matching the printed
// order. A Beacon whose target is gone at resolution fizzles whole
// (CR 608.2b), so it goes to the graveyard rather than the library: the
// shuffle is part of the resolution that did not happen.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "85909caa-d2ad-4487-b9b7-6ac60a14b833",
		Name:         "Beacon of Tomorrows",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := targetPlayerTakesExtraTurns(1)(item, ctx); err != nil {
				return err
			}
			owner := item.Owner
			return ctx.Game.TuckToLibraryThenForEffect(ctx.Source(), game.TuckOptions{}, func(g *game.Game, _ bool) error {
				return g.ShuffleLibraryForEffect(owner)
			})
		},
	})
}
