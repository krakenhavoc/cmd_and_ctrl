package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nessian Asp — Creature — Snake {4}{G}, 4/5 (Theros, #5561):
//
//	"Reach
//	 {6}{G}: Monstrosity 4. (If this creature isn't monstrous, put four +1/+1 counters on it and it becomes monstrous.)"
//
// The shared monstrosity row (CR 701.37).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8b46e077-f667-432f-ab5b-8256a7378311",
		Name:            "Nessian Asp",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated:       []effects.ActivatedAbility{effects.Monstrosity(effects.ManaCost("{6}{G}"), 4)},
	})
}
