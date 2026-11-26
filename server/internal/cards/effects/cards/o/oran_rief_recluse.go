package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oran-Rief Recluse — Creature — Spider {2}{G} 1/3 (Zendikar, S82 OLD batch
// 491, #5583):
//
//	"Kicker {2}{G}
//	 Reach
//	 When this creature enters, if it was kicked, destroy target creature with
//	 flying."
//
// Mold Shambler's shape: an enters trigger whose intervening "if" (CR 603.4)
// reads the permanent's own kicked record as it enters, with the target
// clause naming a creature with flying (read through the layers).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "043fa53c-8a46-4fbd-836f-2403c62e0121",
		Name:            "Oran-Rief Recluse",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		OptionalCosts:   []game.AdditionalCost{effects.Kicker("{2}{G}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedAtLeast(1)),
					"Oran-Rief Recluse — kicked, destroy target creature with flying",
					effects.DestroyFirstLegalCardTarget),
				effects.TargetCreature("target creature with flying", effects.HasKeyword("flying"))),
		},
	})
}
