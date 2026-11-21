package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// News Helicopter — Artifact Creature — Construct {3}, 1/1
// (Marvel's Spider-Man):
//
//	"Flying
//	 When this creature enters, create a 1/1 green and white Human
//	 Citizen creature token."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "192dd1e2-b6c2-432a-a1b4-e52ff39b3db2",
		Name:            "News Helicopter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("News Helicopter — create a 1/1 green and white Human Citizen creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green and white Human Citizen"), N: 1})),
		},
	})
}
