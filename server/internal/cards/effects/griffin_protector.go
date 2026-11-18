package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Griffin Protector — Creature — Griffin {3}{W}, 2/3:
//
//	"Flying
//	 Whenever another creature you control enters, this creature gets
//	 +1/+1 until end of turn."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2d91cf05-fc39-4d76-8c95-62738b57e0a0",
		Name:            "Griffin Protector",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AnotherCreatureEnteredUnderYourControl,
				"Griffin Protector — this creature gets +1/+1 until end of turn",
				thisGetsUntilEndOfTurn(1, 1, "Griffin Protector — +1/+1 until end of turn")),
		},
	})
}
