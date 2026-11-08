package effects

// Sower of Chaos — Creature — Devil {3}{R}, 4/3 (Foundations, #4465):
//
//	"{2}{R}: Target creature can't block this turn."
//
// Brambleback Brute's can't-block row at instant speed and with only a
// mana cost. Any creature is a legal target, the Sower's own included.
//
// No simplification.
func init() {
	const label = "Sower of Chaos — target creature can't block this turn"
	Register(Spec{
		OracleID:     "2ec1860b-89cc-4529-8b8c-fb47f79f2b4b",
		Name:         "Sower of Chaos",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}{R}: Target creature can't block this turn.",
			Cost:    ManaCost("{2}{R}"),
			Targets: TargetCreature("target creature"),
			Purpose: ForTargets(CantBlockUntilEOT(0, 0, 0)),
			Effect:  targetCreatureCantBlockThisTurn(label),
		}},
	})
}
