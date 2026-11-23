package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phyrexian Debaser — Creature — Phyrexian Carrier {3}{B}, 2/2 (Urza's
// Legacy, #5623):
//
//	"Flying
//	 {T}, Sacrifice this creature: Target creature gets -2/-2 until end
//	 of turn."
//
// A Carrier through the shared ulgB1SacrificeToShrink row at -2/-2: {T} and
// sacrificing it are both costs, so it waits out summoning sickness
// (CR 302.6) and is gone before the ability resolves. Flying is a printed
// keyword (CR 702.9b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "554de904-086c-428c-9f32-b0b2498a6fd5",
		Name:            "Phyrexian Debaser",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{
			effects.UlgB1SacrificeToShrink("Phyrexian Debaser", 2),
		},
	})
}
