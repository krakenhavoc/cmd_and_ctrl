package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dictate of Heliod — Enchantment {3}{W}{W} (Journey into Nyx, #5558):
//
//	"Flash
//	 Creatures you control get +2/+2."
//
// Flash, and a +2/+2 anthem over the creatures its controller controls (an
// empty TribeFilter matches every creature).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b989279c-665d-4f15-afc1-adf3872a4851",
		Name:            "Dictate of Heliod",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{YoursOnly: true}, 2, 2),
		},
	})
}
