package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sea Gate Banneret — Creature — Kor Warrior {W}, 1/2 (Zendikar Rising,
// #5508):
//
//	"{4}{W}: Creatures you control get +1/+1 until end of turn."
//
// Timid Shieldbearer's team pump: the creatures are locked in as the
// ability resolves (CR 611.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ddf12758-b356-4ff8-b0ce-547083498335",
		Name:         "Sea Gate Banneret",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{4}{W}: Creatures you control get +1/+1 until end of turn.",
			Cost:    ManaCost("{4}{W}"),
			Purpose: game.Purpose{TeamPump: &game.Pump{Power: 1, Toughness: 1}, Answers: game.AnswerPump},
			Effect: Do(BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 1, Toughness: 1,
				Label: "Sea Gate Banneret — +1/+1"}),
		}},
	})
}
