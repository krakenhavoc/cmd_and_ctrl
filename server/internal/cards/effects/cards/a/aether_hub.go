package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aether Hub — Land:
//
//	"When this land enters, you get {E} (an energy counter).
//	 {T}: Add {C}.
//	 {T}, Pay {E}: Add one mana of any color."
//
// ADR 0129 §5: "{T}, Pay {E}" is still a mana ability (CR 605.1a), with
// an energy component (ManaAbilityCost.Energy) checked before anything
// is paid (CR 118.3). The auto-tapper plans the coloured half in its
// energy tier: only when no plan that spends no energy pays the cost,
// before any plan that pays life, and never for more energy than the
// controller has across the whole plan (owner decision 2). The {C} half
// is an ordinary source.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "61c89b11-65c9-4fda-bbcd-d84de25df801",
		Name:         "Aether Hub",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Energy: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouGetEnergy("Aether Hub", 1),
		},
		ManaAbilities: effects.AnyColorForEnergyRows(1),
	})
}
