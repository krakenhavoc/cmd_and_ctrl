package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Silver Raven — Artifact Creature — Bird {U}, 1/1 (Adventures in the Forgotten Realms):
//
//	Flying
//	 When this creature enters, scry 1.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4da64806-8a1d-4926-a168-695df43309cf",
		Name:            "Silver Raven",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Scry: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Silver Raven — scry 1", effects.Do(effects.Scry{N: 1})),
		},
	})
}
