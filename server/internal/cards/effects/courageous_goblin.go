package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Courageous Goblin — Creature — Goblin {1}{R}, 2/2 (Foundations,
// #4465):
//
//	"Whenever this creature attacks while you control a creature with
//	 power 4 or greater, this creature gets +1/+0 and gains menace until
//	 end of turn."
//
// Nighthowl Pursuer's attack trigger (hobbitFerociousAttack): the
// "while" is read as the attack is declared, and the Goblin itself
// counts if its power is 4 or greater. The pump and menace land on the
// Goblin while it is still on the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c107e271-5741-4a84-bfc4-3bc322972d0d",
		Name:         "Courageous Goblin",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, hobbitFerociousAttack,
				"Courageous Goblin — gets +1/+0 and gains menace until end of turn",
				thisCreatureUntilEOT("Courageous Goblin — +1/+0 and menace", 1, 0, "menace")),
		},
	})
}
