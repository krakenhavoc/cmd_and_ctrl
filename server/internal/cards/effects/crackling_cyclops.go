package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crackling Cyclops — Creature — Cyclops Wizard {2}{R}, 0/4
// (Foundations, #4465):
//
//	"Whenever you cast a noncreature spell, this creature gets +3/+0
//	 until end of turn."
//
// TVA Bureaucrat's trigger with only the pump. The boost lands on the
// Cyclops that triggered; one that has left does nothing (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "16ab9451-f3da-4f61-8e2b-d9c84e0e5428",
		Name:         "Crackling Cyclops",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Noncreature(), "Crackling Cyclops — gets +3/+0 until end of turn",
				thisGetsUntilEndOfTurn(3, 0, "Crackling Cyclops — +3/+0")),
		},
	})
}
