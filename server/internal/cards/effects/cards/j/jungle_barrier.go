package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jungle Barrier — Creature — Plant Wall {2}{G}{U}, 2/6 (Apocalypse,
// #5614):
//
//	"Defender (This creature can't attack.)
//	 When this creature enters, draw a card."
//
// Wall of Blossoms on a bigger body: defender rides PrintedKeywords
// (CR 702.3b), and the enters trigger is a mandatory draw on the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "dcf65293-14a0-46a5-a0b0-ca2b7fda8bef",
		Name:            "Jungle Barrier",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Jungle Barrier — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
