package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:         "11809d9b-30f3-49d5-9119-e45b4de044c8",
		Name:             "Knight Luminary",
		Completeness:     CompletenessFull,
		Purpose:          game.Purpose{Tokens: 1},
		AlternativeCosts: []game.AlternativeCost{Warp("{1}{W}")},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Knight Luminary — create a 1/1 white Human Soldier creature token",
				Do(CreateToken{Template: TokenCard("1/1 white Human Soldier"), N: 1})),
		},
	})
}
