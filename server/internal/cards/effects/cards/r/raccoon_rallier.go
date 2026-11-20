package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Raccoon Rallier — Creature — Raccoon Bard {1}{R}, 2/2 (Bloomburrow,
// #4467):
//
//	"{T}: Target creature you control gains haste until end of turn.
//	 Activate only as a sorcery."
//
// A tap ability, so summoning sickness applies to the Rallier itself
// (the engine enforces it for a creature's {T} cost). The target is
// re-checked as the ability resolves, and the haste lasts for the turn.
//
// No simplification.
func init() {
	const label = "Raccoon Rallier — haste until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "50c038a4-6546-4a15-a117-176849edaa3d",
		Name:         "Raccoon Rallier",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{T}: Target creature you control gains haste until end of turn. Activate only as a sorcery.",
			Cost:         effects.TapCost(),
			Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
			Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "haste")),
			SorcerySpeed: true,
			Effect:       effects.FirstTargetGainsUntilEndOfTurn(0, label, "haste"),
		}},
	})
}
