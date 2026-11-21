package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Celestial Enforcer — Creature — Human Cleric {2}{W}, 2/3 (Foundations
// Jumpstart):
//
//	"{1}{W}, {T}: Tap target creature. Activate only if you control a
//	 creature with flying."
//
// Baxter Building's gate (ControlsAtLeast) over a creature with flying, read
// through the layered view; Arashin Sunshield's tapper body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "81f5c70f-fbb1-4b3a-a048-1f4602adfa0e",
		Name:         "Celestial Enforcer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "{1}{W}, {T}: Tap target creature. Activate only if you control a creature with flying.",
			Cost:      Plus(ManaCost("{1}{W}"), TapCost()),
			Targets:   TargetCreature("target creature"),
			Condition: ControlsAtLeast(1, BnsB27CreatureWithFlying),
			Effect:    TapChosenPermanent,
		}},
	})
}

// bnsB27CreatureWithFlying matches a creature that currently has flying.
func BnsB27CreatureWithFlying(c game.Card) bool {
	return c.IsCreature() && game.HasKeyword(&c, "flying")
}
