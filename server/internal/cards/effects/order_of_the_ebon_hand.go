package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Order of the Ebon Hand — Creature — Cleric Knight {B}{B}, 2/1
// (Fallen Empires, #5637):
//
//	"Protection from white
//	 {B}: This creature gains first strike until end of turn.
//	 {B}{B}: This creature gets +1/+0 until end of turn."
//
// Order of Leitbur's mirror in black: the protection rides
// PrintedKeywords, the two rows change the creature itself until end of
// turn, and neither has {T} in its cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2f65f16f-c6c2-48f0-b556-63cf7d162000",
		Name:            "Order of the Ebon Hand",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"protection from white"},
		Activated: []ActivatedAbility{
			{
				Label:   "{B}: This creature gains first strike until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    ManaCost("{B}"),
				Effect:  thisGainsKeywordUntilEndOfTurn("first strike", "Order of the Ebon Hand — first strike until end of turn"),
			},
			{
				Label:   "{B}{B}: This creature gets +1/+0 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    ManaCost("{B}{B}"),
				Effect:  thisGetsUntilEndOfTurn(1, 0, "Order of the Ebon Hand — +1/+0 until end of turn"),
			},
		},
	})
}
