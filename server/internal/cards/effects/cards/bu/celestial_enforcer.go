package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "81f5c70f-fbb1-4b3a-a048-1f4602adfa0e",
		Name:         "Celestial Enforcer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     "{1}{W}, {T}: Tap target creature. Activate only if you control a creature with flying.",
			Cost:      effects.Plus(effects.ManaCost("{1}{W}"), effects.TapCost()),
			Targets:   effects.TargetCreature("target creature"),
			Condition: effects.ControlsAtLeast(1, bnsB27CreatureWithFlying),
			Effect:    effects.TapChosenPermanent,
		}},
	})
}

// bnsB27CreatureWithFlying matches a creature that currently has flying.
func bnsB27CreatureWithFlying(c game.Card) bool {
	return c.IsCreature() && game.HasKeyword(&c, "flying")
}
