package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flowstone Kavu — Creature — Kavu {2}{R}, 2/3 (Dominaria United, S82):
//
//	"Menace
//	 {R}: This creature gets +1/-1 until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "49475268-be95-4670-a5ca-2f2b6c5cf005",
		Name:            "Flowstone Kavu",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Activated:       []effects.ActivatedAbility{effects.SelfPumpRow("Flowstone Kavu", "{R}", 1, -1)},
	})
}
