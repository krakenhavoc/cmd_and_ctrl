package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Panicked Altisaur — Creature — Dinosaur {4}{R}, 4/5 (The Lost Caverns
// of Ixalan, #4470):
//
//	"Reach
//	 {T}: This creature deals 2 damage to each opponent."
//
// Tal Shiar Analyst's pinger at two damage, with reach as the printed
// keyword. The tap ability is a CR 602 activation with a tap cost, so
// summoning sickness applies (CR 302.6) and the damage uses the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6de64e59-0a47-46c4-90bf-c56e63266f19",
		Name:            "Panicked Altisaur",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated: []ActivatedAbility{{
			Label:   "{T}: This creature deals 2 damage to each opponent.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return damageToEachOpponent(g, item, 2)
			},
		}},
	})
}
