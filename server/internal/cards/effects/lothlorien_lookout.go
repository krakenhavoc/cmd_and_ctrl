package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lothlórien Lookout — Creature — Elf Scout {1}{G}, 1/3 (The Hobbit
// Eternal, HOC):
//
//	"Whenever this creature attacks, scry 1."
//
// The shared attack trigger with a scry for its body (#4454).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ed839fae-3429-4ac5-997c-286d8e8cf05f",
		Name:         "Lothlórien Lookout",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Lothlórien Lookout — scry 1", Do(Scry{N: 1})),
		},
	})
}
