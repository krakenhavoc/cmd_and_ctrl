package cardsh

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

// Healing Leaves — Instant {G} (Planar Chaos, #5595):
//
//	"Choose one —
//	 • Target player gains 3 life.
//	 • Prevent the next 3 damage that would be dealt to any target this turn."
//
// The first bullet gains the target player 3 life; the second is Goblin
// Legionnaire's prevention shield on the bullet's own target, used up by
// the next 3 damage that would be dealt to it this turn (CR 615.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9c675e7a-1d68-44b2-9c1a-543b73782437",
		Name:         "Healing Leaves",
		Completeness: effects.CompletenessFull,
		Modes:        effects.OldB608GainThreeOrPreventThree("Healing Leaves"),
	})
}
