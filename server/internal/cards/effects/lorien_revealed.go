package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "66f28905-c7fc-4ada-8fa0-199626d9bedb",
		Name:         "Lórien Revealed",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 3},
		Activated: []ActivatedAbility{
			Typecycling("Islandcycling", "{1}", "an Island card", IsLandWithSubtype("Island")),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return DrawCards{Player: ctx.Controller(), N: 3}.Apply(ctx)
		},
	})
}
