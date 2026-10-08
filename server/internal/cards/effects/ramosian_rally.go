package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "2731c237-eda1-4b74-a294-502d574f4435",
		Name:         "Ramosian Rally",
		Completeness: CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			TapInstead(1, "an untapped creature you control", ControlsA("Plains"), Creature()),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 1, Toughness: 1, Label: "Ramosian Rally"}.Apply(ctx)
		},
	})
}
