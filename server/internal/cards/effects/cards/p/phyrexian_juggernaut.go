package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phyrexian Juggernaut — Artifact Creature — Phyrexian Juggernaut {6}, 5/5 (Mirrodin Besieged,
// #5576):
//
//	"Infect (This creature deals damage to creatures in the form of
//	 -1/-1 counters and to players in the form of poison counters.)
//	 This creature attacks each combat if able."
//
// Infect is the engine's damage tail (CR 702.90b); the attack requirement
// is Juggernaut's AttacksEachCombat (CR 508.1d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "009fcd1e-5c5f-435b-afc1-099f1622e45c",
		Name:            "Phyrexian Juggernaut",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"infect"},
		Static:          []game.StaticAbility{effects.AttacksEachCombat()},
	})
}
