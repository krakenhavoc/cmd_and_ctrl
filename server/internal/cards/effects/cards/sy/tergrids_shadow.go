package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tergrid's Shadow — Instant {3}{B}{B} (Kaldheim, #5505):
//
//	"Each player sacrifices two creatures.
//	 Foretell {2}{B}{B}"
//
// Barter in Blood's edict (jmpB1EachPlayerSacrificesN: each player
// chooses their own two, in APNAP order, and one with fewer sacrifices
// what they have), plus the shared foretell special action
// (CR 702.143).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "692260bb-e01c-4550-832a-9611856a6ff0",
		Name:           "Tergrid's Shadow",
		Purpose:        game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepSacrifice, Partial: true}},
		Completeness:   effects.CompletenessFull,
		SpecialActions: []game.SpecialAction{effects.Foretell("{2}{B}{B}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.JmpB1EachPlayerSacrificesN(2, "Tergrid's Shadow — choose two creatures to sacrifice")(ctx.Game, item)
		},
	})
}
