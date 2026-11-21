package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Subjugator Angel — Creature — Angel {4}{W}{W}, 4/3 (Innistrad Remastered):
//
//	"Flying
//	 When this creature enters, tap all creatures your opponents control."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "88b1e06c-5899-4e83-8204-e2c32c0c6aff",
		Name:            "Subjugator Angel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Subjugator Angel — tap all creatures your opponents control", effects.TapAllCreaturesYourOpponentsControl),
		},
	})
}
