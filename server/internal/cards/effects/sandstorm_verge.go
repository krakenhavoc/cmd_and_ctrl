package effects

// Sandstorm Verge — Land — Desert (Outlaws of Thunder Junction, #4468):
//
//	"{T}: Add {C}.
//	 {3}, {T}: Target creature can't block this turn. Activate only as a
//	 sorcery."
//
// The colourless mana ability, and Brambleback Brute's sorcery-speed
// "target creature can't block this turn" with {3} and {T} as the
// cost. With the target gone by resolution the ability does nothing.
// Desert is on the printed type line.
//
// No simplification.
func init() {
	const label = "Sandstorm Verge — target creature can't block this turn"
	Register(Spec{
		OracleID:     "de417a82-8f03-4d7e-aee7-48f7d7eba61a",
		Name:         "Sandstorm Verge",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:        "{3}, {T}: Target creature can't block this turn. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{3}"), TapCost()),
			Targets:      TargetCreature("target creature"),
			Purpose:      ForTargets(CantBlockUntilEOT(0, 0, 0)),
			SorcerySpeed: true,
			Effect:       targetCreatureCantBlockThisTurn(label),
		}},
	})
}
