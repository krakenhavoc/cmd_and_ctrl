package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Leonin Sun Standard — Artifact {2} (Mirrodin, #5606):
//
//	"{1}{W}: Creatures you control get +1/+1 until end of turn."
//
// Dwarven Provisioner's team pump: a layer 7c boost on each creature
// its controller controls as it resolves (CR 611.2c), so a creature
// that arrives later this turn gets nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e235f988-55d2-421b-9cfd-0d30d2727aed",
		Name:         "Leonin Sun Standard",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}: Creatures you control get +1/+1 until end of turn.",
			Purpose: game.Purpose{TeamPump: &game.Pump{Power: 1, Toughness: 1}, Answers: game.AnswerPump},
			Cost:    ManaCost("{1}{W}"),
			Effect: Do(BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 1, Toughness: 1,
				Label: "Leonin Sun Standard — creatures you control get +1/+1"}),
		}},
	})
}
