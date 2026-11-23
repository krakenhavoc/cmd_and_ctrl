package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// DOT-7 Repair Squad — {3}{W} Artifact Creature — Robot 3/3 (Star Trek,
// #4973):
//
//	"Flying
//	 {1}{W}: Untap target creature. Activate only as a sorcery and only
//	 once each turn."
//
// The untap is untapTheTarget on any creature, the Squad included. Both
// timing limits are CR 602.5d restrictions read at announce:
// SorcerySpeed for "as a sorcery", and OncePerTurnActivation for "only
// once each turn", which counts this object's activations of the
// ability (CR 602.1b), so a countered activation is still the turn's one.
//
// No simplification.
func init() {
	const label = "{1}{W}: Untap target creature. Activate only as a sorcery and only once each turn."
	effects.Register(effects.Spec{
		OracleID:        "b860f64c-e830-49ea-a478-5dc87e8a3be1",
		Name:            "DOT-7 Repair Squad",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:        label,
			Cost:         effects.ManaCost("{1}{W}"),
			SorcerySpeed: true,
			Condition:    effects.OncePerTurnActivation(label),
			Targets:      effects.TargetCreature("target creature"),
			Effect:       effects.UntapTheTarget,
		}},
	})
}
