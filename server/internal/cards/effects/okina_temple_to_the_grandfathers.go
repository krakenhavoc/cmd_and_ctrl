package effects

// Okina, Temple to the Grandfathers — Legendary Land (Champions of
// Kamigawa, #5603):
//
//	"{T}: Add {G}.
//	 {G}, {T}: Target legendary creature gets +1/+1 until end of turn."
//
// Shizo, Death's Storehouse's shape with Sword of the Chosen's body: the
// bonus ends in the cleanup step (CR 514.2), and a target that stops
// being legendary in response gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3ad69bfb-2e51-4fe9-8d2f-7d071a4f1c69",
		Name:         "Okina, Temple to the Grandfathers",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{G}, {T}: Target legendary creature gets +1/+1 until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 1, 1)),
			Cost:    Plus(ManaCost("{G}"), TapCost()),
			Targets: TargetCreature("target legendary creature", Legendary()),
			Effect:  boostTargetUntilEOT("Okina, Temple to the Grandfathers — +1/+1", 1, 1),
		}},
	})
}
