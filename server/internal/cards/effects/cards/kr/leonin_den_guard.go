package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Leonin Den-Guard — Creature — Cat Soldier {1}{W}, 1/3 (Mirrodin, #5606):
//
//	"As long as this creature is equipped, it gets +1/+1 and has
//	 vigilance."
//
// Armed Assailant's "as long as it's equipped" (mrdB5WhileEquippedPlusOne):
// a layer 7c boost and a layer 6 grant on the creature itself, both
// read live from the attachment relation (CR 301.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "651220f7-223a-49b3-995d-32f08eb3a16a",
		Name:         "Leonin Den-Guard",
		Completeness: effects.CompletenessFull,
		Static:       effects.MrdB5WhileEquippedPlusOne("vigilance"),
	})
}
