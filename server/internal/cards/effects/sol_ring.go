package effects

// Sol Ring — "{T}: Add {C}{C}." Colorless Medallion-tier artifact.
//
// Fully implemented. The ManaAbilities slot below is live: the S15
// mana pipeline resolves it off-stack per CR 605.3a and puts {C}{C}
// in the controller's pool.
//
// The S14 note this file used to carry — "mana abilities are
// activated abilities, and the pipeline lands in S19 … mana is
// tracked on paper today" — outlived its fix by several sprints and
// was corrected in the #338 stale-simplification sweep.
func init() {
	Register(Spec{
		OracleID: "6ad8011d-3471-4369-9d68-b264cc027487",
		Name:     "Sol Ring",
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}{C}",
			Label:    "Add {C}{C}",
		}},
	})
}
