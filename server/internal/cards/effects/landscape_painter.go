package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Landscape Painter // Vibrant Idea — Creature — Merfolk Wizard {1}{U},
// 2/1 // Sorcery {4}{U} (preparation card, CR 722):
//
//	"This creature enters prepared. (While it's prepared, you may cast a
//	 copy of its spell. Doing so unprepares it.)"
//
//	Vibrant Idea — "Draw two cards."
//
// The SORCERY-speed proof card for ADR 0090. The copy of Vibrant Idea
// in exile carries no timing of its own and the derived permission
// grants none, so it waits for its controller's main phase with an
// empty stack like any sorcery.
//
// No simplification.
const landscapePainterOracleID = "5ba7abbc-f6e8-40c2-800b-11efd10cca5b"

func init() {
	Register(Spec{
		OracleID:     landscapePainterOracleID,
		Name:         "Landscape Painter",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersPrepared()},
	})
	Register(Spec{
		OracleID:     landscapePainterOracleID + "#1",
		Name:         "Vibrant Idea",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DrawCards{N: 2}.Apply(ctx)
		},
	})
}
