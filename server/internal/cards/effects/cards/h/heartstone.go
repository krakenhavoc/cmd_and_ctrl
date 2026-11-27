package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heartstone — Artifact {3} (Stronghold, S82 OLD b751, #5627):
//
//	"Activated abilities of creatures cost {1} less to activate. This effect
//	 can't reduce the mana in that cost to less than one mana."
//
// Training Grounds' activation-scoped reduction (ActivationCostsLess with the
// one-mana floor, #302) at {1}, over every creature on the battlefield rather
// than only the controller's: a cost of {1} or no mana is left alone, a {2}
// ability costs {1}, and an ability that works from another zone (cycling) is
// not one of a creature. Mana abilities are activated abilities (CR 605.1a), so
// theirs are included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8f8ebf49-37c5-4da0-9cee-84ad22ab5d37",
		Name:         "Heartstone",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.KeepsOneMana(effects.ActivationCostsLess(1,
				"Activated abilities of creatures cost {1} less to activate. This effect can't reduce the mana in that cost to less than one mana.",
				func(q game.CostQuery) bool {
					return q.FromZone == game.ZoneBattlefield && q.Card.IsCreature()
				})),
		},
	})
}
