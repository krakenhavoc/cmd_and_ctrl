package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blood Servitor — Artifact Creature — Construct {3}, 2/2 (Innistrad: Crimson Vow):
//
//	"When this creature enters, create a Blood token."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5a3f9c4d-20a8-46aa-9cf3-3916f56a7906",
		Name:         "Blood Servitor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Blood Servitor — create a Blood token", effects.Do(effects.S82Vow03CreateBlood(1))),
		},
	})
}
