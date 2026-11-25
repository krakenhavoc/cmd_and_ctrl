package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lightning Visionary — Creature — Minotaur Shaman {1}{R}, 2/1 (Jumpstart,
// #5510):
//
//	"Prowess (Whenever you cast a noncreature spell, this creature gets
//	 +1/+1 until end of turn.)"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3ad5af12-1741-427d-92c4-6a730cd4faa2",
		Name:            "Lightning Visionary",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"prowess"},
	})
}
