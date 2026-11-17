package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Silkenfist Fighter — Creature — Kor Soldier {1}{W}, 1/3
// (Nemesis, #5618):
//
//	"Whenever this creature becomes blocked, untap it."
//
// The trigger is nemUntapWhenBlocked: it triggers once however many
// creatures block (CR 509.3c), and the untapped creature stays
// attacking and blocked.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bf5ceba2-9def-4ff4-9291-740ca90c1c4d",
		Name:         "Silkenfist Fighter",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{nemUntapWhenBlocked("Silkenfist Fighter")},
	})
}
