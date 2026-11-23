package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ridged Kusite — Creature — Horror Spellshaper {B}, 1/1 (Planar Chaos, #5595):
//
//	"{1}{B}, {T}, Discard a card: Target creature gets +1/+0 and gains first
//	 strike until end of turn."
//
// pumpAndGrantFirstTarget on an activated row whose cost discards: the
// still-legal target (CR 608.2b) gets the boost and the keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a0c3c26e-e9e8-424d-922e-a692bb479ebd",
		Name:         "Ridged Kusite",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}, {T}, Discard a card: Target creature gets +1/+0 and gains first strike until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0, "first strike")),
			Cost:    effects.Plus(effects.ManaCost("{1}{B}"), effects.TapCost(), effects.DiscardACard()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.PumpAndGrantFirstTarget(1, 0, "Ridged Kusite — +1/+0 and first strike until end of turn", "first strike"),
		}},
	})
}
