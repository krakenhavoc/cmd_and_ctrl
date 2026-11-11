package effects

// Glittering Stockpile — Artifact — Treasure {2}{R} (Outlaws of Thunder
// Junction Commander, #5368):
//
//	"{T}: Add {R}. Put a stash counter on this artifact.
//	 {T}, Sacrifice this artifact: Add X mana of any one color, where X is
//	 the number of stash counters on this artifact."
//
// Two mana abilities on a Treasure card (the subtype is printed, so
// Treasure payoffs see it). The first adds {R} and then puts the counter on
// (a Rider: the counter follows the mana). The second is the Treasure's
// sacrifice with a scaled amount: the sacrifice is a cost, so X is the
// counters the artifact had as it left the battlefield, and the mana is
// all of one color, picked once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "feaf9dcb-d845-4cf9-b35d-cd58175162e7",
		Name:         "Glittering Stockpile",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{R}",
				Rider:    cmdr2S33PutAStashCounter,
				Label:    "{T}: Add {R}. Put a stash counter on this artifact",
			},
			{
				Cost:           ManaAbilityCost{Tap: true, Sacrifice: true},
				ProducedAmount: ProducedOneColor(cmdr2S33StashCounters),
				Label:          "{T}, Sacrifice this artifact: Add X mana of any one color, where X is the number of stash counters on this artifact",
			},
		},
	})
}
