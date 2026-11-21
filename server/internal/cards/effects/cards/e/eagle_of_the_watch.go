package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eagle of the Watch — Creature — Bird {2}{W}, 2/1 (Journey into Nyx, #5558):
//
//	"Flying, vigilance"
//
// Armored Griffin's printed keywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "954836f8-c225-450a-bb51-f2adb6923ef5",
		Name:            "Eagle of the Watch",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
	})
}
