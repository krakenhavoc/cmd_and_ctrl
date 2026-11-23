package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// City of Shadows — Land (#1600):
//
//	"{T}, Exile a creature you control: Put a storage counter on this
//	 land.
//	 {T}: Add {C} for each storage counter on this land."
//
// The storage land whose counter is bought with a creature. The first
// ability adds no mana, so it is an ordinary activated ability that
// uses the stack (CR 605.1a); its cost is the exile-a-permanent
// component (game.ExilePermanentsCost, ADR 0020's 2026-10-03
// amendment), named at announce and paid before the ability is on the
// stack. The creature is exiled, not sacrificed: leaves-the-battlefield
// triggers see it, dies and sacrifice triggers do not.
//
// The second is a mana ability that counts the storage counters as it
// is activated (ProducedPerCounterOnThis) and leaves them on the land —
// unlike Mage-Ring Network, City of Shadows never spends its bank. With
// no counters it taps for nothing, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "043192d8-6077-46c3-b43f-b7caf6762869",
		Name:         "City of Shadows",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:           effects.ManaAbilityCost{Tap: true},
			ProducedAmount: effects.ProducedPerCounterOnThis(game.CounterStorage, "C"),
			Label:          "Add {C} for each storage counter on this land",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Exile a creature you control: Put a storage counter on this land.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.TapCost(), effects.ExileACreatureYouControl()),
			Effect:  effects.PutStorageCounterOnThis,
		}},
	})
}
