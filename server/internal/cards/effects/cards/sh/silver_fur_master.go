package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Silver-Fur Master — Creature — Rat Ninja {U}{B}, 2/2:
//
//	"Ninjutsu {U}{B} ({U}{B}, Return an unblocked attacker you control
//	 to hand: Put this card onto the battlefield from your hand tapped
//	 and attacking.)
//	 Ninjutsu abilities you activate cost {1} less to activate.
//	 Other Ninja and Rogue creatures you control get +1/+1."
//
// The discount is a board cost modifier scoped to activations, matched
// on the ninjutsu keyword bit (ANinjutsuAbilityCost, #5672), so it
// reaches a ninjutsu or commander ninjutsu ability (CR 702.49d) its
// controller activates from any card, printed or granted. It is a
// static ability of a permanent, so it does nothing for this card's own
// ninjutsu while the card is in hand, and it takes off generic mana
// only (CR 118.7a): its own {U}{B} is never cheaper.
//
// The lord is "creatures you control that are Ninjas or Rogues, other
// than this one", each pumped once.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "60ec735d-c9be-4776-beb7-f603ae22edbf",
		Name:         "Silver-Fur Master",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Ninjutsu("{U}{B}")},
		CostModifiers: []game.CostModifier{
			effects.ActivationCostsLess(1, "Ninjutsu abilities you activate cost {1} less to activate.",
				effects.ANinjutsuAbilityCost(), effects.ActivatedByTheModifiersController()),
		},
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Ninja", "Rogue"}, Others: true, YoursOnly: true}, 1, 1),
		},
	})
}
