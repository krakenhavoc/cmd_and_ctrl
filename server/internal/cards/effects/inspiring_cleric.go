package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Inspiring Cleric — Creature — Vampire Cleric {2}{W}, 3/2 (Ixalan, #5532):
//
//	"When this creature enters, you gain 4 life."
//
// Spiritual Guardian's body: an enters trigger on the stack (CR 603.6a)
// that gains its controller 4 life.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "65e19aec-e8ea-412f-a239-bd8d1b78249c",
		Name:         "Inspiring Cleric",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Inspiring Cleric — you gain 4 life", Do(GainLife{Amount: 4})),
		},
	})
}
