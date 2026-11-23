package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thrashing Mossdog — Creature — Plant Dog {3}{G}, 3/3 (Ravnica: Clue Edition, #5929): "Reach. Scavenge {4}{G}{G}."
//
// Scavenge is the shared graveyard ability (Deadbridge Goliath).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8778ddbe-9aba-4a45-a22a-5453fefef952",
		Name:            "Thrashing Mossdog",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated:       []effects.ActivatedAbility{effects.Scavenge("{4}{G}{G}")},
	})
}
