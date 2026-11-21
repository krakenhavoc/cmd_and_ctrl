package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Imperious Oligarch — Creature — Human Cleric {W}{B}, 2/1 (Ravnica
// Allegiance, #5521):
//
//	"Vigilance
//	 Afterlife 1 (When this creature dies, create a 1/1 white and black
//	 Spirit creature token with flying.)"
//
// Vigilance is printed data (CR 702.20b). Afterlife is the dies trigger it
// reminds (CR 702.135a), Syndicate Messenger's 1/1 white and black Spirit
// with flying.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b0381c2a-40e2-4002-82e3-ae2d687daed3",
		Name:            "Imperious Oligarch",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Purpose:         game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Imperious Oligarch — afterlife 1: create a 1/1 white and black Spirit creature token with flying",
				effects.KnightOfTheLastBreathSpirits(1)),
		},
	})
}
