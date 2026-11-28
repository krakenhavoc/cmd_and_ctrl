package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wayfaring Giant — Creature — Giant {5}{W}, 1/3 (Invasion, #5616):
//
//	"Domain — This creature gets +1/+1 for each basic land type among lands
//	 you control."
//
// A layer 7c static (S82OldB660SelfPumpDomain), its count read on every pass.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e20e5ad9-36b5-4962-acc2-555d069c5daf",
		Name:         "Wayfaring Giant",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.S82OldB660SelfPumpDomain(1, 1, "Domain — This creature gets +1/+1 for each basic land type among lands you control."),
		},
	})
}
