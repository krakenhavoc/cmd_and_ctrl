package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Order of Leitbur — Creature — Human Cleric Knight {W}{W}, 2/1
// (Fallen Empires, #5637):
//
//	"Protection from black
//	 {W}: This creature gains first strike until end of turn.
//	 {W}{W}: This creature gets +1/+0 until end of turn."
//
// The protection rides PrintedKeywords; the two rows are Boros
// Reckoner's first strike and Shivan Dragon's pump on the creature
// itself. Neither has {T} in its cost, so both can be activated any
// number of times, summoning sickness or not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d8f5279f-c58e-4560-9c68-77d226a3c86c",
		Name:            "Order of Leitbur",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"protection from black"},
		Activated: []ActivatedAbility{
			{
				Label:   "{W}: This creature gains first strike until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    ManaCost("{W}"),
				Effect:  thisGainsKeywordUntilEndOfTurn("first strike", "Order of Leitbur — first strike until end of turn"),
			},
			{
				Label:   "{W}{W}: This creature gets +1/+0 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    ManaCost("{W}{W}"),
				Effect:  thisGetsUntilEndOfTurn(1, 0, "Order of Leitbur — +1/+0 until end of turn"),
			},
		},
	})
}
