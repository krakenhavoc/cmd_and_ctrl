package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wispmare — Creature — Elemental {2}{W}, 1/3 (Lorwyn, S82 OLD, #5593):
//
//	"Flying
//	 When this creature enters, destroy target enchantment.
//	 Evoke {W}"
//
// Ronom Unicorn's destroy of the enchantment still legal as the trigger
// resolves, under Mulldrifter's evoke.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "a3bb8ce0-e61e-438d-8875-32473bb5b61c",
		Name:             "Wispmare",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{effects.Evoke("{W}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Wispmare — destroy target enchantment", effects.DestroyFirstLegalTarget),
				effects.TargetPermanent("target enchantment", effects.Enchantment())),
		},
	})
}
