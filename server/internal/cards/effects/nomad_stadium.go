package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(odyThresholdLand("4034bec6-e3c7-4d3f-81df-7c903977a606", "Nomad Stadium", "{W}",
		"You gain 4 life.", ActivatedAbility{
			Purpose: game.Purpose{Answers: game.AnswerValue, LifeGain: 4},
			Effect:  Do(GainLife{Amount: 4}),
		}))
}
