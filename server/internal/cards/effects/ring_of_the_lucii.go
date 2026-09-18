package effects

// Ring of the Lucii — Legendary Artifact for {4} (EDHREC rank 4435):
//
//	"{T}: Add {C}{C}.
//	 {2}, {T}, Pay 1 life: Tap target nonland permanent."
//
// A four-mana rock that also locks something down every turn — and
// because the two abilities share the tap, each turn is a choice
// between the mana and the tap-down. Roadmap batch 42 (#449), "no new
// machinery".
//
// The two abilities are different KINDS and the file is the place
// that distinction shows. "Add {C}{C}" is a mana ability: it does not
// use the stack (CR 605.3), resolves synchronously the moment it is
// activated, and cannot be responded to. The tap-down is an ordinary
// CR 602 activated ability that goes on the stack with its target
// chosen at announce, so it can be answered.
//
// The life is a COST, not a rider: it is validated before anything is
// paid, so an activation at 0 life is rejected outright and the Ring
// does not tap. The {2} is mana that has to be floating first, as the
// engine's activated-ability cost path requires.
//
// "Target NONLAND permanent" — the Ring cannot tap down a land, which
// is what keeps it out of the mana-denial slot it would otherwise
// occupy.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d37e595f-afe7-4ad9-87db-b9ceb3f435e8",
		Name:         "Ring of the Lucii",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}{C}",
			Label:    "Add {C}{C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}, Pay 1 life: Tap target nonland permanent.",
			Cost:    Plus(ManaCost("{2}"), TapCost(), PayLife(1)),
			Targets: TargetPermanent("target nonland permanent", Nonland()),
			Effect:  tapChosenPermanent,
		}},
	})
}
