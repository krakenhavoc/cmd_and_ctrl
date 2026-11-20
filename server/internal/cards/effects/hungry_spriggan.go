package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hungry Spriggan — Creature — Goblin Warrior {2}{G}, 1/1 (Shadowmoor,
// #5591):
//
//	"Trample
//	 Whenever this creature attacks, it gets +3/+3 until end of turn."
//
// Trample is a printed keyword. The attack trigger is Lurking
// Nightstalker's at +3/+3 until end of turn (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f80c8533-b29c-49ae-a241-ef4b73a9e327",
		Name:            "Hungry Spriggan",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Hungry Spriggan — it gets +3/+3 until end of turn",
				thisGetsUntilEndOfTurn(3, 3, "Hungry Spriggan — +3/+3")),
		},
	})
}
