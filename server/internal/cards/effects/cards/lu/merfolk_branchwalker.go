package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Merfolk Branchwalker — Creature — Merfolk Scout {1}{G}, 2/1:
//
//	"When this creature enters, it explores. (Reveal the top card of your
//	 library. Put that card into your hand if it's a land. Otherwise, put a
//	 +1/+1 counter on this creature, then put the card back or put it into
//	 your graveyard.)"
//
// The engine's explore action (CR 701.44) with the Branchwalker as the
// explorer.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "24b5dbe3-578a-4897-8215-e723f00a8f1a",
		Name:         "Merfolk Branchwalker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Merfolk Branchwalker — it explores",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					return effects.Explores{Explorer: ctx.Source()}.Apply(ctx)
				}))),
		},
	})
}
