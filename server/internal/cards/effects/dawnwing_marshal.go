package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dawnwing Marshal — Creature — Cat Soldier {1}{W}, 2/2 (Foundations,
// #4465):
//
//	"Flying
//	 {4}{W}: Creatures you control get +1/+1 until end of turn."
//
// Bearer of Glory's pump on a flier: flying rides PrintedKeywords, and
// the pump reaches the creatures you control as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6640c234-8863-4670-a61f-ba6f59b60d69",
		Name:            "Dawnwing Marshal",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{4}{W}: Creatures you control get +1/+1 until end of turn.",
			Cost:    ManaCost("{4}{W}"),
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Effect: Do(BoostUntilEOT{
				Match: And(Creature(), YouControl()), Power: 1, Toughness: 1,
				Label: "Dawnwing Marshal — creatures you control get +1/+1",
			}),
		}},
	})
}
