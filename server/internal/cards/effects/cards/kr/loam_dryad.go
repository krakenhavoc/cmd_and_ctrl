package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Loam Dryad — Creature — Dryad Horror {G}, 1/2
// (Shadows over Innistrad, #5543):
//
//	"{T}, Tap an untapped creature you control: Add one mana of any
//	 color."
//
// Citanul Stalwart's mana ability with a creature filter. The {T} taps the
// Dryad, so the other creature is another one (TapAnotherUntapped); it is
// not the {T} symbol (CR 302.6), so one that arrived this turn pays it,
// and it does not target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1420b7db-859b-4ad5-9844-10413cc70478",
		Name:         "Loam Dryad",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true, TapOthers: effects.TapAnotherUntapped("an untapped creature you control", effects.Creature()).TapOthers},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
	})
}
