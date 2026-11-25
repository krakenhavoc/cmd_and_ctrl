package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// New Perspectives — {5}{U} Enchantment (Commander 2020):
//
//	"When this enchantment enters, draw three cards.
//	 As long as you have seven or more cards in hand, you may pay {0} rather
//	 than pay cycling costs."
//
// The draw is the entry trigger. The static is a CR 118.9 alternative cost for
// cycling and typecycling abilities (Spec.CyclingCostAlternatives), the
// equip-cost alternative's twin-row machinery with the discard of the cycled
// card kept: the card being cycled is still in hand as the cost is announced,
// so it counts toward the seven (CR 601.2b/602.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e1b8a89a-b308-4fc5-ada7-61de6e21b7f9",
		Name:         "New Perspectives",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 3},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("New Perspectives — draw three cards", effects.Do(effects.DrawCards{N: 3})),
		},
		CyclingCostAlternatives: []game.EquipCostAlternative{{
			Label: "As long as you have seven or more cards in hand, you may pay {0} rather than pay cycling costs.",
			Cost:  "{0}",
			Condition: func(g *game.Game, controller uuid.UUID) bool {
				p := g.PlayerByIDForEffect(controller)
				return p != nil && p.Hand != nil && p.Hand.Size() >= 7
			},
		}},
	})
}
