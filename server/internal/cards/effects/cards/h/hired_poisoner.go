package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hired Poisoner — Creature — Human Assassin {B}, 1/1 (Guilds of Ravnica, #5525):
//
//	"Deathtouch"
//
// Deathtouch is an engine keyword carried as printed data (CR 702.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a7a6cdfd-10ac-49d0-a0a6-8ad37e326511",
		Name:            "Hired Poisoner",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
	})
}
