package effects

// Blinkmoth Well — Land (Mirrodin, #5606):
//
//	"{T}: Add {C}.
//	 {2}, {T}: Tap target noncreature artifact."
//
// Witch's Clinic's colourless mana ability beside Auriok Transfixer's
// tapper, on a noncreature-artifact clause: an artifact creature (an
// animated Blinkmoth Nexus, an artifact land that is a creature) is not
// a legal target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dd3fd7cd-d8a1-48e2-861d-c7c40086dc75",
		Name:         "Blinkmoth Well",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}: Tap target noncreature artifact.",
			Cost:    Plus(ManaCost("{2}"), TapCost()),
			Targets: TargetPermanent("target noncreature artifact", Artifact(), Noncreature()),
			Effect:  tapChosenPermanent,
		}},
	})
}
