package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Experiment Twelve — Creature — Elf Lizard Warrior {3}{G}, 4/4 (Murders at
// Karlov Manor Commander, #5368):
//
//	"Trample
//	 Whenever this creature or another creature you control is turned face
//	 up, put +1/+1 counters on that creature equal to its power.
//	 Disguise {4}{G}"
//
// The permanent is face up, with its own power, by the time the event is
// emitted (CR 708.8), so the counters equal the power it has now.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "67a10da3-76f2-4f4a-90d1-e77abb5842fe",
		Name:             "Experiment Twelve",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"trample"},
		AlternativeCosts: []game.AlternativeCost{Disguise("{4}{G}")},
		Triggered: []game.TriggeredAbility{
			On(game.EventTurnedFaceUp, ACreatureYouControlWasTurnedFaceUp,
				"Experiment Twelve — put +1/+1 counters on that creature equal to its power",
				cmdr2S38PutCountersEqualToPower),
		},
	})
}
