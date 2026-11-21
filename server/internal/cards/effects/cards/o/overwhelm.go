package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Overwhelm — Sorcery {5}{G}{G} (Foundations Jumpstart):
//
//	"Convoke (Your creatures can help cast this spell. Each creature you
//	 tap while casting this spell pays for {1} or one mana of that
//	 creature's color.)
//	 Creatures you control get +3/+3 until end of turn."
//
// Overcome's team pump without the trample: the set of creatures is
// pinned as the spell resolves (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "caeaad5e-2d68-481c-9a35-c2c3b345963d",
		Name:         "Overwhelm",
		Completeness: effects.CompletenessFull,
		TapCost:      effects.Convoke(),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 3, Toughness: 3, Label: "Overwhelm — +3/+3"}.Apply(ctx)
			}),
		),
	})
}
