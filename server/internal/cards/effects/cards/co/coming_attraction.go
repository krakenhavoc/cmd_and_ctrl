package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "5d9c7ca4-f735-4bee-b03f-4c0c76ee48ed",
		Name:         "Coming Attraction",
		Purpose:      game.Purpose{Lands: 1},
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.SearchLibrary{
					Player:        ctx.Controller(),
					Predicate:     effects.IsBasicLand,
					Dest:          game.ZoneBattlefield,
					Limit:         1,
					Reveal:        true,
					Shuffle:       true,
					TappedOnEntry: true,
					Reason:        "Coming Attraction — a basic land",
				}.Apply(ctx)
			}),
			effects.OpenAttractions{N: 1},
		),
	})
}
