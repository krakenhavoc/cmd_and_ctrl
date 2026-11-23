package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rakshasa Deathdealer — Creature — Demon {B}{G}, 2/2 (Khans of Tarkir,
// #5554):
//
//	"{B}{G}: This creature gets +2/+2 until end of turn.
//	 {B}{G}: Regenerate this creature."
//
// Two shared rows: selfPumpRow (the firebreathing shape, any number of
// times) and regenerateThisRow (a regeneration shield, CR 701.19a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "acdc17cf-8f8c-42ee-bede-481ce00cf4bf",
		Name:         "Rakshasa Deathdealer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.SelfPumpRow("Rakshasa Deathdealer", "{B}{G}", 2, 2),
			effects.RegenerateThisRow("{B}{G}"),
		},
	})
}
