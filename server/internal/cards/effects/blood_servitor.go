package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blood Servitor — Artifact Creature — Construct {3}, 2/2 (Innistrad: Crimson Vow):
//
//	"When this creature enters, create a Blood token."
func init() {
	Register(Spec{
		OracleID:     "5a3f9c4d-20a8-46aa-9cf3-3916f56a7906",
		Name:         "Blood Servitor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Blood Servitor — create a Blood token", Do(s82Vow03CreateBlood(1))),
		},
	})
}
