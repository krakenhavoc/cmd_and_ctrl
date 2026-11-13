package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Silkenfist Order — Creature — Kor Soldier {3}{W}{W}, 3/5
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
		OracleID:     "5cca974b-1a6c-485f-adc0-81d2b7dca286",
		Name:         "Silkenfist Order",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{nemUntapWhenBlocked("Silkenfist Order")},
	})
}
