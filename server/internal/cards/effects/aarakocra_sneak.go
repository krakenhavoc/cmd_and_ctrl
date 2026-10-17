package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aarakocra Sneak — Creature — Bird Rogue {3}{U}, 1/4:
//
//	"Flying
//	 When this creature enters, you take the initiative."
//
// Flying is engine-side; the ETB is TakeTheInitiative (ADR 0203,
// #4096), whose venture into Undercity is the initiative's own trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c1882cb4-f69e-441e-8871-743ed636cad0",
		Name:            "Aarakocra Sneak",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouTakeTheInitiative("Aarakocra Sneak"),
		},
	})
}
