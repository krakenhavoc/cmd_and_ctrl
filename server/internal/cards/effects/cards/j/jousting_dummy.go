package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jousting Dummy — Artifact Creature — Scarecrow Knight {2}, 2/1 (Throne of Eldraine, #5516):
//
//	"{3}: This creature gets +1/+0 until end of turn."
//
// The shared pump row (thisGetsRow).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "96eea161-8af0-452f-b992-576efdf58d88",
		Name:         "Jousting Dummy",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.ThisGetsRow("{3}", 1, 0, "Jousting Dummy")},
	})
}
