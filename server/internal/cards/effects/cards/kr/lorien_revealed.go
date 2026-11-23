package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lórien Revealed — Sorcery {3}{U}{U} (The Hobbit):
//
//	"Draw three cards.
//	 Islandcycling {1} ({1}, Discard this card: Search your library for
//	 an Island card, reveal it, put it into your hand, then shuffle.)"
//
// A plain draw-three, plus typecycling (CR 702.29e) whose search takes any
// land with the Island subtype.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "66f28905-c7fc-4ada-8fa0-199626d9bedb",
		Name:         "Lórien Revealed",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 3},
		Activated: []effects.ActivatedAbility{
			effects.Typecycling("Islandcycling", "{1}", "an Island card", effects.IsLandWithSubtype("Island")),
		},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DrawCards{Player: ctx.Controller(), N: 3}.Apply(ctx)
		},
	})
}
