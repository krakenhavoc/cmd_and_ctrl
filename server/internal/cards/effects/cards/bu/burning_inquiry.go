package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "af98156b-3064-4f16-940e-10039241f2b0",
		Name:         "Burning Inquiry",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.B33EachPlayerDrawsThenDiscardsAtRandom(ctx, 3)
		},
	})
}
