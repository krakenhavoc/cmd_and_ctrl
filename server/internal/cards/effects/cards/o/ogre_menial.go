package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ogre Menial — Creature — Phyrexian Ogre {3}{R}, 0/4 (New Phyrexia, #5575):
//
//	"Infect (This creature deals damage to creatures in the form of -1/-1 counters and to players in the form of poison counters.)
//	 {R}: This creature gets +1/+0 until end of turn."
//
// Infect (CR 702.90) and the shared firebreathing row (thisGetsRow).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e7c879fa-a34b-42d9-8edf-29254380c1f7",
		Name:            "Ogre Menial",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"infect"},
		Activated:       []effects.ActivatedAbility{effects.ThisGetsRow("{R}", 1, 0, "Ogre Menial")},
	})
}
