package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Altar of the Pantheon — Artifact {3} (Theros Beyond Death, #5514):
//
//	"Your devotion to each color and each combination of colors is increased by one.
//	 {T}: Add one mana of any color. If you control a God, a Demigod, or a legendary
//	 enchantment, you gain 1 life."
//
// The devotion clause is the devotion-bonus seam (RegisterDevotionBonus, read by
// DevotionToColors, so every devotion count, a God's, Gray Merchant's, a cost
// reduction's, sees it): +1 once per Altar for any colour or colours (CR 700.5).
// The mana ability's rider reads the board as the ability resolves (CR 605.3a):
// the life is gained after the mana is added, if the condition holds then.
//
// No simplification.
func init() {
	const oracle = "17d41bbf-6a84-48c4-baac-e45f8ed86b78"
	effects.RegisterDevotionBonus(oracle, 1)
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Altar of the Pantheon",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: effects.OneColorOfAmount(1),
			Label:    "Add one mana of any color. If you control a God, a Demigod, or a legendary enchantment, you gain 1 life",
			Rider: func(g *game.Game, controller, source uuid.UUID) error {
				if !effects.ControlsAGodDemigodOrLegendaryEnchantment(g, controller, source) {
					return nil
				}
				return effects.GainLife{Player: controller, Amount: 1}.Apply(effects.KeyedContext(g, controller, source))
			},
		}},
	})
}
