package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Maw of Kozilek — Creature — Eldrazi Drone {3}{R}, 2/5 (Oath of the Gatewatch, #5545):
//
//	"Devoid (This card has no color.)
//	 {C}: This creature gets +2/-2 until end of turn. ({C} represents
//	 colorless mana.)"
//
// Devoid makes it colourless (CR 702.114a). The +2/-2 is the shared
// thisGetsRow (Darklit Gargoyle's +2/-1 shape). {C} is colorless mana
// only, never generic (CR 107.4c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "11b529f0-e365-4500-b54a-44e2ac6e709f",
		Name:            "Maw of Kozilek",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Activated:       []ActivatedAbility{thisGetsRow("{C}", 2, -2, "Maw of Kozilek")},
	})
}
