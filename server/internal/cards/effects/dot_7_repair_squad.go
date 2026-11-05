package effects

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
	Register(Spec{
		OracleID:        "b860f64c-e830-49ea-a478-5dc87e8a3be1",
		Name:            "DOT-7 Repair Squad",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:        label,
			Cost:         ManaCost("{1}{W}"),
			SorcerySpeed: true,
			Condition:    OncePerTurnActivation(label),
			Targets:      TargetCreature("target creature"),
			Effect:       untapTheTarget,
		}},
	})
}
