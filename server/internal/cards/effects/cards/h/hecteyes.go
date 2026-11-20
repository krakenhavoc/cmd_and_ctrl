package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hecteyes — Creature — Ooze Horror {1}{B}, 1/1 (Final Fantasy, #4462):
//
//	"When this creature enters, each opponent discards a card."
//
// Arbiter of Woe's discard (eachOpponentDiscardsOne): each opponent
// chooses a card without revealing it and the cards are discarded at
// the same time.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7a11b4ed-430d-4ef8-b0c1-6a022f12448c",
		Name:         "Hecteyes",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Hecteyes — each opponent discards a card",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error { return effects.EachOpponentDiscardsOne(ctx.Game, ctx.Item) }))),
		},
	})
}
