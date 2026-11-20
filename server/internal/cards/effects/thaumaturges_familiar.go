package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thaumaturge's Familiar — Artifact Creature — Bird {3}, 1/3 (Theros Beyond Death, #5514):
//
//	"Flying
//	 When this creature enters, scry 1."
//
// Glider Kids' shape: flying and an enters scry.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "74a1a484-7c35-466a-b097-9aaf6fc2073b",
		Name:            "Thaumaturge's Familiar",
		Completeness:    CompletenessFull,
		Purpose:         game.Purpose{Scry: 1},
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Thaumaturge's Familiar — scry 1", Do(Scry{N: 1})),
		},
	})
}
