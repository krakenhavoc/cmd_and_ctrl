package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grafted Skullcap — Artifact {4} (Urza's Saga, S82 OLD b722):
//
//	At the beginning of your draw step, draw an additional card.
//	At the beginning of your end step, discard your hand.
//
// Two triggers: an extra draw in your draw step (Midnight Oil's shape) and a
// discard of the whole hand at your end step.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7694887a-2f3b-4ef1-b598-f3323f018ef4",
		Name:         "Grafted Skullcap",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.On(game.EventBeginDrawStep, effects.ByYou, "Grafted Skullcap — draw an additional card", effects.Do(effects.DrawCards{N: 1})),
				game.Purpose{Draws: 1}),
			effects.AtYourEndStep("Grafted Skullcap — discard your hand", effects.S82OldB722DiscardYourHand),
		},
	})
}
