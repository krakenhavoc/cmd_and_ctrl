package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Willow-Wind — Creature — Elemental {4}{U}, 3/4 (The Hobbit Eternal, HOC):
//
//	"Flying
//	 When this creature enters, scry 2."
//
// Faerie Seer's text on a bigger body (#4454).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ee3fd92c-48e0-4651-aeda-7eb2c75acbb0",
		Name:            "Willow-Wind",
		Purpose:         game.Purpose{Scry: 2},
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Willow-Wind — scry 2", effects.Do(effects.Scry{N: 2})),
		},
	})
}
