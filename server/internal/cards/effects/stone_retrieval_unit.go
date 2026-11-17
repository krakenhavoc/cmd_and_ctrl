package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "31d9a771-09ae-4cd4-bd19-6012f839a9e8",
		Name:         "Stone Retrieval Unit",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Stone Retrieval Unit — create a tapped Powerstone token",
				Do(CreateTokenAdvanced{Spec: Token(PowerstoneToken()).EntersTapped(), N: 1})),
		},
	})
}
