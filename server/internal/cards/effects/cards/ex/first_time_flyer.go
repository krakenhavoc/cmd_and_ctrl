package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// First-Time Flyer — Creature — Human Pilot Ally {1}{U}, 1/2 (Avatar:
// The Last Airbender, #4459):
//
//	"Flying
//	 This creature gets +1/+1 as long as there's a Lesson card in your
//	 graveyard."
//
// The pump is a layer 7c self-modifier read live on every layer pass,
// so it switches on the moment a Lesson reaches the graveyard and off
// the moment the last one leaves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cbbea4e2-4b36-4965-89b6-00d2a564f146",
		Name:            "First-Time Flyer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{effects.AvB4LessonSelfPT(1, 1)},
	})
}
