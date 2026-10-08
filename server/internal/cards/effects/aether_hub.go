package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "61c89b11-65c9-4fda-bbcd-d84de25df801",
		Name:         "Aether Hub",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Energy: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouGetEnergy("Aether Hub", 1),
		},
		ManaAbilities: anyColorForEnergyRows(1),
	})
}
