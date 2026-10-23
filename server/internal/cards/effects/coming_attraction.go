package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Coming Attraction — Sorcery {2}{G}:
//
//	"Search your library for a basic land card, put it onto the
//	 battlefield tapped, then shuffle. Open an Attraction. (Put the top
//	 card of your Attraction deck onto the battlefield.)"
//
// Rampant Growth's search, then the open. The open waits for the search
// (CR 608.2c): the steps run in order and the second holds while the
// first asks which basic.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5d9c7ca4-f735-4bee-b03f-4c0c76ee48ed",
		Name:         "Coming Attraction",
		Purpose:      game.Purpose{Lands: 1},
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return SearchLibrary{
					Player:        ctx.Controller(),
					Predicate:     IsBasicLand,
					Dest:          game.ZoneBattlefield,
					Limit:         1,
					Reveal:        true,
					Shuffle:       true,
					TappedOnEntry: true,
					Reason:        "Coming Attraction — a basic land",
				}.Apply(ctx)
			}),
			OpenAttractions{N: 1},
		),
	})
}
