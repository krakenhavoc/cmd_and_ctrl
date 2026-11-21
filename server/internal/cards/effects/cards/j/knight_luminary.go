package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Knight Luminary — Creature — Human Knight {3}{W}, 3/2 (Edge of
// Eternities, #4461):
//
//	"When this creature enters, create a 1/1 white Human Soldier creature
//	 token.
//	 Warp {1}{W}"
//
// Bastion of Remembrance's entry token with Mechanozoa's warp offer. The
// Soldier is a separate object: when a warped Knight is exiled at the
// next end step, the token stays (CR 702.185a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "11809d9b-30f3-49d5-9119-e45b4de044c8",
		Name:             "Knight Luminary",
		Completeness:     effects.CompletenessFull,
		Purpose:          game.Purpose{Tokens: 1},
		AlternativeCosts: []game.AlternativeCost{effects.Warp("{1}{W}")},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Knight Luminary — create a 1/1 white Human Soldier creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Human Soldier"), N: 1})),
		},
	})
}
