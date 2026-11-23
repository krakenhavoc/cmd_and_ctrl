package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Seraph of the Suns — Creature — Angel {5}{W}{W}, 4/4
// (Amonkhet, #5535):
//
//	"Flying
//	 Indestructible (Damage and effects that say "destroy" don't destroy this creature. If its toughness is 0 or less, it still dies.)"
//
// Printed keywords only, as Kaldra Compleat declares indestructible.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8354b70e-43fc-4581-bb53-b913335bf460",
		Name:            "Seraph of the Suns",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "indestructible"},
	})
}
