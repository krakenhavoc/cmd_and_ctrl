package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Furtive Homunculus — "Skulk (This creature can't be blocked by creatures
// with greater power.)"
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6b59ffa2-5c8b-4549-aa09-2f2a5cabc5af",
		Name:            "Furtive Homunculus",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"skulk"},
	})
}
