package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Viashino Sandscout — Creature — Lizard Scout {1}{R}, 2/1 (Urza's
// Legacy, #5623):
//
//	"Haste (This creature can attack and {T} as soon as it comes under
//	 your control.)
//	 At the beginning of the end step, return this creature to its
//	 owner's hand. (Return it only if it's on the battlefield.)"
//
// Viashino Sandstalker's shape through the shared ulgB1ReturnAtEachEndStep
// trigger: "the end step" is any player's, so one put onto the battlefield
// on another turn goes back at that turn's end, and only while it is still
// the same object (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f8d5a65c-382b-4277-9c08-5dec9e395b02",
		Name:            "Viashino Sandscout",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			ulgB1ReturnAtEachEndStep("Viashino Sandscout"),
		},
	})
}
