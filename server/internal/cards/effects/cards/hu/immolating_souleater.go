package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Immolating Souleater — Artifact Creature — Phyrexian Dog {2}, 1/1 (New Phyrexia, #5575):
//
//	"{R/P}: This creature gets +1/+0 until end of turn. ({R/P} can be paid with either {R} or 2 life.)"
//
// The shared firebreathing row (thisGetsRow); the Phyrexian symbol is
// paid with {R} or 2 life at activation (CR 107.4f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fb260f5a-b937-435b-ab18-dcc57d6845df",
		Name:         "Immolating Souleater",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.ThisGetsRow("{R/P}", 1, 0, "Immolating Souleater")},
	})
}
