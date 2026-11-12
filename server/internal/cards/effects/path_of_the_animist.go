package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Path of the Animist — Sorcery {3}{G} (March of the Machine Commander,
// #5648):
//
//	"Search your library for up to two basic land cards, put them onto
//	 the battlefield tapped, then shuffle.
//	 Will of the Planeswalkers — Starting with you, each player votes for
//	 planeswalk or chaos. If planeswalk gets more votes, planeswalk. If
//	 chaos gets more votes or the vote is tied, chaos ensues."
//
// Explosive Vegetation's search, then the vote (WillOfThePlaneswalkers),
// which waits for the search to be answered (CR 608.2c). The vote happens
// at every table; planeswalking and chaos do something only at a table
// playing Planechase.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "85ccae95-92c4-41ff-96fd-dcb515dc568e",
		Name:         "Path of the Animist",
		Purpose:      game.Purpose{Lands: 2},
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return SearchLibrary{
					Player:        ctx.Controller(),
					Predicate:     IsBasicLand,
					Dest:          game.ZoneBattlefield,
					Limit:         2,
					Reveal:        true,
					Shuffle:       true,
					TappedOnEntry: true,
					Reason:        "Path of the Animist — up to two basic lands",
				}.Apply(ctx)
			}),
			WillOfThePlaneswalkers{Card: "Path of the Animist"},
		),
	})
}
