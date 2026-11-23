package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thieves' Fortune — Kindred Instant — Rogue {2}{U} (Morningtide, #5592):
//
//	"Prowl {U} (You may cast this for its prowl cost if you dealt combat
//	 damage to a player this turn with a Rogue.)
//	 Look at the top four cards of your library. Put one of them into your
//	 hand and the rest on the bottom of your library in any order."
//
// Impulse's sentence (LookAtTopThenTakeOneRestOnBottom) with prowl, the
// gated alternative cost of prowl_keyword.go (CR 702.76a), reading the
// Kindred card's Rogue type.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "e0bffcf3-8d34-461b-a1e4-ae47c1e7b19e",
		Name:             "Thieves' Fortune",
		Completeness:     effects.CompletenessFull,
		Purpose:          game.Purpose{Draws: 1},
		AlternativeCosts: []game.AlternativeCost{effects.Prowl("{U}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.LookAtTopThenTakeOneRestOnBottom(4, "Thieves' Fortune — put one of them into your hand")(ctx.Game, item)
		},
	})
}
