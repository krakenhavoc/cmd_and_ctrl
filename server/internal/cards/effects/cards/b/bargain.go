package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bargain — {2}{W} Sorcery (Starter 1999, S82 OLD batch 702, #5620):
//
//	"Target opponent draws a card.
//	 You gain 7 life."
//
// The opponent draws first, then you gain the life; with the target gone the
// spell does nothing (CR 608.2b; S82OldB702Bargain).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a0dd88f6-6e36-40ce-bac2-a0db2b0117b6",
		Name:         "Bargain",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve:    effects.S82OldB702Bargain,
	})
}
