package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Warrior's Oath — Sorcery {R}{R}:
//
//	"Take an extra turn after this one. At the beginning of that turn's
//	 end step, you lose the game."
//
// Last Chance's text under another name; see final_fortune.go.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "574044cf-2e2f-4b5c-b4d1-cf05ba814ab2",
		Name:         "Warrior's Oath",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.ExtraTurnThenLoseAtItsEndStep("Warrior's Oath — you lose the game"),
	})
}
