package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Warren Elder — Creature — Rabbit Cleric {1}{W}, 2/2 (Bloomburrow,
// #4467):
//
//	"{3}{W}: Creatures you control get +1/+1 until end of turn."
//
// Dwarven Provisioner's ability: the creatures you control as it
// resolves, locked in for the turn (CR 611.2c), so one that enters later
// gets nothing. No {T} in the cost, so it can be activated again, and
// while the Elder is summoning sick.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "33fcb05b-246d-4b82-8b0f-44d5f891b612",
		Name:         "Warren Elder",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}{W}: Creatures you control get +1/+1 until end of turn.",
			Purpose: game.Purpose{TeamPump: &game.Pump{Power: 1, Toughness: 1}, Answers: game.AnswerPump},
			Cost:    ManaCost("{3}{W}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 1, Toughness: 1,
					Label: "Warren Elder — creatures you control get +1/+1"}.Apply(NewContext(g, item))
			},
		}},
	})
}
