package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sway of Illusion — Instant {1}{U} (Invasion, S82 #5616):
//
//	"Any number of target creatures become the color of your choice until end of
//	 turn.
//	 Draw a card."
//
// Distorting Lens's layer 5 colour change, on any number of creatures: one
// colour is chosen as the spell resolves and replaces every colour each
// creature still legal as a target had (CR 613.1e), until cleanup.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9a7e2298-8855-43a7-8cb1-4b31e4058c3f",
		Name:         "Sway of Illusion",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      effects.TargetCreature("any number of target creatures").WithCount(0, 0),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.SwayOfIllusionResolve(ctx.Game, item)
		},
	})
}
