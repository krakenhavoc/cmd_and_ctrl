package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fiend Binder — Creature — Human Soldier {3}{W}, 3/2 (Eldritch Moon,
// #5541):
//
//	"Whenever this creature attacks, tap target creature defending player
//	 controls."
//
// Distracting Geist's attack trigger (distractingGeistTap), word for word.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "060c7999-ab81-4e37-af65-7cd219b693c8",
		Name:         "Fiend Binder",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{distractingGeistTap("Fiend Binder — tap target creature defending player controls")},
	})
}
