package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rigging Runner — Creature — Goblin Pirate {R}, 1/1 (Jumpstart 2022, #5948):
//
//	"First strike
//	 Raid — This creature enters with a +1/+1 counter on it if you attacked
//	 this turn."
//
// Swaggering Corsair's raid: a self-replacement on the entry (CR 614.1c),
// read as the Runner enters against the turn's attack record.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "72b600e3-d671-4279-8ca6-9a3d6e774b53",
		Name:            "Rigging Runner",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Replacements:    []game.ReplacementEffect{avB16SelfEntersWithCountersIfRaid(game.CounterPlusOne, 1)},
	})
}
