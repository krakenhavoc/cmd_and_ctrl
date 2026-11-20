package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Necrogen Scudder — Creature — Phyrexian Horror {2}{B}, 3/3 (Scars of Mirrodin, #5577):
//
//	"Flying
//	 When this creature enters, you lose 3 life."
//
// Foul Imp's shape: printed flying and an enters trigger that loses the
// life (youLoseLife).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "93f37df7-e1c4-457b-bc22-00c49b0c4f79",
		Name:            "Necrogen Scudder",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Necrogen Scudder — you lose 3 life", Do(youLoseLife(3))),
		},
	})
}
