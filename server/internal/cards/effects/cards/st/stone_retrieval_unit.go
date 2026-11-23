package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stone Retrieval Unit — Artifact Creature — Construct {4}, 2/3 (The
// Brothers' War, #5482):
//
//	"When this creature enters, create a tapped Powerstone token."
//
// The Powerstone enters tapped because this ability says so (the
// 2022-10-14 ruling), through CreateTokenAdvanced's EntersTapped spec,
// the shape Stern Lesson uses.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "31d9a771-09ae-4cd4-bd19-6012f839a9e8",
		Name:         "Stone Retrieval Unit",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Stone Retrieval Unit — create a tapped Powerstone token",
				effects.Do(effects.CreateTokenAdvanced{Spec: effects.Token(effects.PowerstoneToken()).EntersTapped(), N: 1})),
		},
	})
}
