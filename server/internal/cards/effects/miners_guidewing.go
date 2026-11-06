package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Miner's Guidewing — Creature — Bird {W}, 1/1 (The Lost Caverns of
// Ixalan, #4470):
//
//	"Flying, vigilance
//	 When this creature dies, target creature you control explores."
//
// Flying and vigilance are printed keywords (CR 702.9a, CR 702.20a). The
// death trigger looks back in time (CR 603.10a) and targets a creature
// its controller controls as it goes on the stack; the Guidewing itself
// is gone by then. The explore is the Map token's body
// (targetCreatureYouControlExplores): a land on top goes to the hand,
// anything else gives the creature a +1/+1 counter and asks whether the
// card goes to the graveyard (CR 701.44a). A target gone by resolution
// does not explore (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "eb161d0b-e534-463c-b1d1-7210bfba5d28",
		Name:            "Miner's Guidewing",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisDies("Miner's Guidewing — target creature you control explores",
				targetCreatureYouControlExplores),
				targetCreatureYouControl()),
		},
	})
}
