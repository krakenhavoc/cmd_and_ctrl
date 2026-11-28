package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stunted Growth — Sorcery {3}{G}{G} (Ice Age, S82 OLD, #5636):
//
//	"Target player chooses three cards from their hand and puts them on top of their library in any order."
//
// Brainstorm's put-back for the target player, with three cards: they pick which, then the order. A hand of fewer
// than three goes back whole, and an empty hand is not asked.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e3480c8f-15ad-451b-bb34-b129805569e1",
		Name:         "Stunted Growth",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve:    effects.S82OldB844StuntedGrowth,
	})
}
