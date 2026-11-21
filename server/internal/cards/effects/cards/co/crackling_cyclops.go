package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "16ab9451-f3da-4f61-8e2b-d9c84e0e5428",
		Name:         "Crackling Cyclops",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Noncreature(), "Crackling Cyclops — gets +3/+0 until end of turn",
				effects.ThisGetsUntilEndOfTurn(3, 0, "Crackling Cyclops — +3/+0")),
		},
	})
}
