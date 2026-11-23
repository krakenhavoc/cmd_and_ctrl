package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nomad Stadium — Land (Odyssey, #5613):
//
//	"{T}: Add {W}. This land deals 1 damage to you.
//	 Threshold — {W}, {T}, Sacrifice this land: You gain 4 life. Activate
//	 only if there are seven or more cards in your graveyard."
//
// Odyssey's white threshold land, built by odyThresholdLand. The life
// gain has no target, so nothing can make it fizzle.
//
// No simplification.
func init() {
	effects.Register(effects.OdyThresholdLand("4034bec6-e3c7-4d3f-81df-7c903977a606", "Nomad Stadium", "{W}",
		"You gain 4 life.", effects.ActivatedAbility{
			Purpose: game.Purpose{Answers: game.AnswerValue, LifeGain: 4},
			Effect:  effects.Do(effects.GainLife{Amount: 4}),
		}))
}
