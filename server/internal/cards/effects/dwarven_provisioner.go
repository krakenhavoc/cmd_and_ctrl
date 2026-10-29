package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dwarven Provisioner — Creature — Dwarf Citizen {1}{W}, 2/2 (The Hobbit):
//
//	"{3}{W}: Creatures you control get +1/+1 until end of turn."
//
// The creatures are the ones you control as the ability resolves, locked
// in for the turn (CR 611.2c): one that arrives later gets nothing. It has
// no {T} in its cost, so it can be activated again and again, and with
// summoning sickness.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8fcab221-8737-4357-b8ee-bad9421739a3",
		Name:         "Dwarven Provisioner",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}{W}: Creatures you control get +1/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{3}{W}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 1, Toughness: 1,
					Label: "Dwarven Provisioner — creatures you control get +1/+1"}.Apply(NewContext(g, item))
			},
		}},
	})
}
