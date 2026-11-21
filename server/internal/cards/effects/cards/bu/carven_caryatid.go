package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Carven Caryatid — Creature — Spirit {1}{G}{G}, 2/5 (Tarkir: Dragonstorm
// Commander):
//
//	"Defender
//	 When this creature enters, draw a card."
//
// Defender rides PrintedKeywords; the draw is Baleful Strix's enters
// trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bddf5e9a-7c41-45a8-8fa1-18093c408e15",
		Name:            "Carven Caryatid",
		Completeness:    effects.CompletenessFull,
		Purpose:         game.Purpose{Draws: 1},
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Carven Caryatid — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
