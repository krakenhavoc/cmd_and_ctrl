package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "6640c234-8863-4670-a61f-ba6f59b60d69",
		Name:            "Dawnwing Marshal",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{W}: Creatures you control get +1/+1 until end of turn.",
			Cost:    effects.ManaCost("{4}{W}"),
			Purpose: game.Purpose{TeamPump: &game.Pump{Power: 1, Toughness: 1}, Answers: game.AnswerPump},
			Effect: effects.Do(effects.BoostUntilEOT{
				Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Toughness: 1,
				Label: "Dawnwing Marshal — creatures you control get +1/+1",
			}),
		}},
	})
}
