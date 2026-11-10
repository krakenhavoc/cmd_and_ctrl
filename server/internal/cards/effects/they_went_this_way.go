package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// They Went This Way — Sorcery {2}{G} (Murders at Karlov Manor, #4469):
//
//	"Search your library for a basic land card, put it onto the
//	 battlefield tapped, then shuffle. Investigate. (Create a Clue token.
//	 It's an artifact with '{2}, Sacrifice this token: Draw a card.')"
//
// Thunderherd Migration's search without the reveal (the text puts the
// card onto the battlefield, not into a hand), then one Clue
// (CR 701.16a). A search that finds nothing still shuffles and still
// investigates.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "47951720-9584-4ca3-a73d-3d259102b2f8",
		Name:         "They Went This Way",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Lands: 1},
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return SearchLibrary{
					Player:        ctx.Controller(),
					Predicate:     IsBasicLand,
					Dest:          game.ZoneBattlefield,
					Limit:         1,
					Shuffle:       true,
					TappedOnEntry: true,
					Reason:        "They Went This Way — a basic land, onto the battlefield tapped",
				}.Apply(ctx)
			}),
			Investigate{N: 1},
		),
	})
}
