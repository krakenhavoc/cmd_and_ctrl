package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Springmane Cervin — Creature — Elk {2}{G}, 3/2 (Strixhaven: School of
// Mages, #5502):
//
//	"When this creature enters, you gain 2 life."
//
// An enters trigger on the stack over a plain life gain.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4e964f31-c764-4af5-8c27-a46fd721fab4",
		Name:         "Springmane Cervin",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Springmane Cervin — you gain 2 life", Do(GainLife{Amount: 2})),
		},
	})
}
