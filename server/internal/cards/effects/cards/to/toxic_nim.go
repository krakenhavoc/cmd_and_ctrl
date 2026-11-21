package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Toxic Nim — Creature — Phyrexian Zombie {4}{B}{B}, 4/1 (New Phyrexia, #5575):
//
//	"Infect (This creature deals damage to creatures in the form of -1/-1 counters and to players in the form of poison counters.)
//	 {B}: Regenerate this creature."
//
// Blight Mamba's shape: infect (CR 702.90) and the shared regenerate row
// (CR 701.19a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "afc2a3a7-eb69-4fcb-b3a3-444dccb735d7",
		Name:            "Toxic Nim",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"infect"},
		Activated:       []effects.ActivatedAbility{effects.RegenerateThisRow("{B}")},
	})
}
