package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rampaging Growth — Instant {3}{G} (Jumpstart 2022, #5948):
//
//	"Search your library for a basic land card, put it onto the battlefield,
//	 then shuffle. Until end of turn, that land becomes a 4/3 Insect creature
//	 with reach and haste. It's still a land."
//
// A basic-land fetch (untapped) whose tail animates the land that was found
// (Cmdr2B13Animate's layered effect: Creature and Insect added in layer 4,
// reach and haste in layer 6, base 4/3 in layer 7b, all pinned to that
// object). A search that finds nothing animates nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "acd81d2c-e438-44db-86eb-32fc9cb3fd8f",
		Name:         "Rampaging Growth",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue, Lands: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.GapB171RampagingGrowth(ctx.Game, item)
		},
	})
}
