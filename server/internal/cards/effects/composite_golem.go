package effects

// Composite Golem — Artifact Creature — Golem {6}, 4/4 (Fifth Dawn, #5604):
//
//	"Sacrifice this creature: Add {W}{U}{B}{R}{G}."
//
// Blood Pet's sacrifice mana ability, adding one mana of each colour.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "784970de-ef8c-4672-b5d0-24a4b0a978a3",
		Name:         "Composite Golem",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Sacrifice: true},
			Produced: "{W}{U}{B}{R}{G}",
			Label:    "Sacrifice this creature: Add {W}{U}{B}{R}{G}",
		}},
	})
}
