package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Corrosion — Enchantment {1}{B}{R} (Visions, S82 older sets batch 789,
// #5631):
//
//	"Cumulative upkeep {1}
//	 At the beginning of your upkeep, put a rust counter on each artifact
//	 target opponent controls. Then destroy each artifact with mana value less
//	 than or equal to the number of rust counters on it. Artifacts destroyed
//	 this way can't be regenerated.
//	 When this enchantment leaves the battlefield, remove all rust counters
//	 from all permanents."
//
// Rust is a counter with no rule of its own (CR 122.1). The destruction looks
// at every artifact on the battlefield, whoever controls it, and an artifact
// with mana value 0 dies at once. The leaves-the-battlefield trigger clears
// every permanent's rust counters, including ones another Corrosion put there.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3d3fd8a2-78d4-4f52-8fcd-0c6fa93f0043",
		Name:         "Corrosion",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepArtifacts, How: game.SweepDestroy, Partial: true}},
		Triggered: []game.TriggeredAbility{
			effects.CumulativeUpkeep("Corrosion — cumulative upkeep {1}", "{1}"),
			effects.Targeting(effects.AtYourUpkeep("Corrosion — put a rust counter on each artifact target opponent controls, then destroy each artifact with mana value no greater than its rust counters",
				effects.Old789CorrosionUpkeep), effects.TargetPlayer("target opponent", effects.Opponent())),
			effects.WhenThisLeaves("Corrosion — remove all rust counters from all permanents", effects.Old789RemoveAllRustCounters),
		},
	})
}
