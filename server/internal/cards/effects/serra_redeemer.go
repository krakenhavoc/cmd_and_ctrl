package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Serra Redeemer — Creature — Angel Soldier {3}{W}{W}, 2/4 (Bloomburrow,
// #4467):
//
//	"Flying
//	 Whenever another creature you control with power 2 or less enters,
//	 put two +1/+1 counters on that creature."
//
// Neighborhood Guardian's trigger condition (the power read as the
// creature enters), with Geralf, the Fleshwright's counters on the
// creature that entered. The counters go on it only while it is the
// same object on the battlefield: one that has left by resolution gets
// none (CR 400.7, CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "39bb051f-1ca7-4d9b-b33b-7f559f816e3e",
		Name:            "Serra Redeemer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AnotherCreatureWithPower2OrLessEnteredUnderYourControl,
				"Serra Redeemer — put two +1/+1 counters on that creature", serraRedeemerCounters),
		},
	})
}

// serraRedeemerCounters puts two +1/+1 counters on the creature whose
// entry triggered the ability, if it is still on the battlefield.
func serraRedeemerCounters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	info, ok := ctx.TriggeringPermanent()
	if !ok || info.Left {
		return nil
	}
	return AddCounter{Target: ctx.Trigger().Object.ID, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
}
