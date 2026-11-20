package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Phyrexian Crusader — Creature — Phyrexian Zombie Knight {1}{B}{B}, 2/2 (Mirrodin Besieged,
// #5576):
//
//	"First strike, protection from red and from white
//	 Infect (This creature deals damage to creatures in the form of
//	 -1/-1 counters and to players in the form of poison counters.)"
//
// First strike, protection from red and from white (two qualities, CR
// 702.16a) and infect, the engine's damage tail (CR 702.90b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "91760bdf-a82a-4f3d-b925-3bc81ffbd41f",
		Name:            "Phyrexian Crusader",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike", game.ProtectionFromColor("R"), game.ProtectionFromColor("W"), "infect"},
	})
}
