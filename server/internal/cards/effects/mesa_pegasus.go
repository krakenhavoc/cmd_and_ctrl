package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mesa Pegasus — Creature — Pegasus {1}{W}, 1/1 (Limited Edition Alpha):
//
//	"Flying; banding"
//
// Flying and banding are both the engine's (game/banding.go, ADR 0224).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "8161f5b8-6aab-4133-ba2c-2e7b5774153e",
		Name:            "Mesa Pegasus",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", game.KeywordBanding},
	})
}
