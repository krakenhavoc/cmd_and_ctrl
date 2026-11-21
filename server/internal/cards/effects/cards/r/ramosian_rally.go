package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ramosian Rally — Instant {3}{W}:
//
//	"If you control a Plains, you may tap an untapped creature you
//	 control rather than pay this spell's mana cost.
//	 Creatures you control get +1/+1 until end of turn."
//
// ADR 0135 §1 (#2030): the tap alternative cost (CR 118.9), offered only
// while you control a Plains. The creatures are the ones you control as
// the spell resolves (CR 611.2c), the tapped one included.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2731c237-eda1-4b74-a294-502d574f4435",
		Name:         "Ramosian Rally",
		Completeness: effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			effects.TapInstead(1, "an untapped creature you control", effects.ControlsA("Plains"), effects.Creature()),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Toughness: 1, Label: "Ramosian Rally"}.Apply(ctx)
		},
	})
}
