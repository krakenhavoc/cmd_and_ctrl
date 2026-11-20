package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Viashino Sandsprinter — Creature — Lizard Warrior {1}{R}{R}, 4/1 (Modern Horizons, #5519):
//
//	"Trample, haste
//	 At the beginning of the end step, return this creature to its owner's
//	 hand. (Return it only if it's on the battlefield.)
//	 Cycling {R} ({R}, Discard this card: Draw a card.)"
//
// Viashino Sandscout's shape (the shared end-step return, at every
// player's end step) with trample and Skycrash's cycling.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6560517d-41c0-4c9c-b9d5-3ecfcb2b99fd",
		Name:            "Viashino Sandsprinter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Triggered: []game.TriggeredAbility{
			ulgB1ReturnAtEachEndStep("Viashino Sandsprinter"),
		},
		Activated: []ActivatedAbility{Cycling("{R}")},
	})
}
