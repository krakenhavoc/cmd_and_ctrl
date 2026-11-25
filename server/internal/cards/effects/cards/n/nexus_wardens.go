package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nexus Wardens — Creature — Satyr Archer {2}{G}, 1/4:
//
//	"Reach
//	 Constellation — Whenever an enchantment you control enters, you gain 2
//	 life."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a2652158-e63f-477b-b328-cae7ef2263bd",
		Name:            "Nexus Wardens",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.GapB103EnchantmentEnteredUnderYourControl,
				"Nexus Wardens — you gain 2 life (constellation)", effects.Do(effects.GainLife{Amount: 2})),
		},
	})
}
