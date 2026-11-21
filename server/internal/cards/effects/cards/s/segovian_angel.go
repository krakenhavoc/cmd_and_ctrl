package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Segovian Angel — "Flying, vigilance."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "50372c54-359c-4ee6-9e21-1e40ba53652d",
		Name:            "Segovian Angel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
	})
}
