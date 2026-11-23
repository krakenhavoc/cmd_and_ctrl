package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lothlórien Lookout — Creature — Elf Scout {1}{G}, 1/3 (The Hobbit
// Eternal, HOC):
//
//	"Whenever this creature attacks, scry 1."
//
// The shared attack trigger with a scry for its body (#4454).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ed839fae-3429-4ac5-997c-286d8e8cf05f",
		Name:         "Lothlórien Lookout",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Lothlórien Lookout — scry 1", effects.Do(effects.Scry{N: 1})),
		},
	})
}
