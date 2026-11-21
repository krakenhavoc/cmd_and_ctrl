package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Last Chance — Sorcery {R}{R}:
//
//	"Take an extra turn after this one. At the beginning of that turn's
//	 end step, you lose the game."
//
// Final Fortune at sorcery speed; see final_fortune.go.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "360039a5-1cbd-4ee3-8f94-21b5348e106a",
		Name:         "Last Chance",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.ExtraTurnThenLoseAtItsEndStep("Last Chance — you lose the game"),
	})
}
