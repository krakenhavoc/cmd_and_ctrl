package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blessed Reversal — Instant {1}{W} (Eighth Edition, #5607):
//
//	"You gain 3 life for each creature attacking you."
//
// Arachnogenesis' count of the creatures attacking the player itself, read as
// the spell resolves: a creature attacking one of your planeswalkers is not
// attacking you.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "70e0b676-61a2-4dcf-8f61-d9281467ed43",
		Name:         "Blessed Reversal",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.OldB608BlessedReversal,
	})
}
