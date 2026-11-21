package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wall of Lost Thoughts — Creature — Wall {1}{U}, 0/4 (Jumpstart):
//
//	"Defender
//	 When this creature enters, target player mills four cards."
//
// A plain enters trigger whose target player, any player you included, is
// chosen as it goes on the stack; the body is TargetPlayerMills. A library
// holding fewer than four mills what it has, and milling out never loses
// the game (CR 701.17b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "49b80368-d38a-4877-b1de-f3b5f3d40d4d",
		Name:            "Wall of Lost Thoughts",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Wall of Lost Thoughts — target player mills four cards", effects.TargetPlayerMills(4)),
				effects.TargetPlayer("target player")),
		},
	})
}
