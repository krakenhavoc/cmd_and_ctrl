package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cylian Elf — {1}{G} Creature — Elf Scout 2/2 (Shards of Alara, S82 OLD b513):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see grizzly_bears.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2d425f5d-326f-46f8-ab84-852b01db4adb",
		Name:         "Cylian Elf",
		Completeness: effects.CompletenessFull,
	})
}
