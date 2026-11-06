package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Icecave Crasher — Creature — Beast {3}{G}, 4/4 (Edge of Eternities,
// #4461):
//
//	"Trample
//	 Landfall — Whenever a land you control enters, this creature gets
//	 +1/+0 until end of turn."
//
// Attercop's landfall pump with +1/+0: one land, one trigger, each
// lasting until end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b9650a0e-5795-42ff-84c6-6040e5cf4c90",
		Name:            "Icecave Crasher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			Landfall("Icecave Crasher — +1/+0 until end of turn (landfall)",
				thisGetsUntilEndOfTurn(1, 0, "Icecave Crasher — +1/+0 until end of turn")),
		},
	})
}
