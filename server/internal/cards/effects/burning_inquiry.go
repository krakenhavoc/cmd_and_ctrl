package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Burning Inquiry — Sorcery {R} (EDHREC rank 3470):
//
//	"Each player draws three cards, then discards three cards at
//	 random."
//
// The one-mana wheel that fills every graveyard. Every player draws
// before anyone discards, in APNAP order, and the discards are the
// engine's random discard — no prompt, no choice, which is the
// printed card. A player with fewer than three cards after the draw
// discards what they have.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "af98156b-3064-4f16-940e-10039241f2b0",
		Name:         "Burning Inquiry",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b33EachPlayerDrawsThenDiscardsAtRandom(ctx, 3)
		},
	})
}
