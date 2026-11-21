package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "47951720-9584-4ca3-a73d-3d259102b2f8",
		Name:         "They Went This Way",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Lands: 1},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.SearchLibrary{
					Player:        ctx.Controller(),
					Predicate:     effects.IsBasicLand,
					Dest:          game.ZoneBattlefield,
					Limit:         1,
					Shuffle:       true,
					TappedOnEntry: true,
					Reason:        "They Went This Way — a basic land, onto the battlefield tapped",
				}.Apply(ctx)
			}),
			effects.Investigate{N: 1},
		),
	})
}
