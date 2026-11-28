package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Delusions of Mediocrity — Enchantment {3}{U} (Urza's Legacy, S82 OLD b716, #5623):
//
//	"When this enchantment enters, you gain 10 life.
//	When this enchantment leaves the battlefield, you lose 10 life."
//
// Two triggers on the stack (Wormfang Manta's enters/leaves pair): leaving is
// any exit, and the life is lost by whoever controlled the enchantment as it
// left (CR 603.10a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a1fb6dfc-1dab-4337-8311-b4ab2ef336c0",
		Name:         "Delusions of Mediocrity",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Delusions of Mediocrity — you gain 10 life", effects.Do(effects.GainLife{Amount: 10})),
			effects.WhenThisLeaves("Delusions of Mediocrity — you lose 10 life", effects.Do(effects.YouLoseLife(10))),
		},
	})
}
