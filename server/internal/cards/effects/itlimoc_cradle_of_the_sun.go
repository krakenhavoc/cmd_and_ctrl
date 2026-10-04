package effects

// Itlimoc, Cradle of the Sun — Legendary Land, the back face of
// Growing Rites of Itlimoc (growing_rites_of_itlimoc.go):
//
//	"(Transforms from Growing Rites of Itlimoc.)
//	 {T}: Add {G}.
//	 {T}: Add {G} for each creature you control."
//
// Two mana abilities; the second is Gaea's Cradle's, counted as it is
// activated. It is a land, not a creature, so it taps for mana the
// turn it transforms (CR 302.6 is about creatures).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     growingRitesOfItlimocOracleID + "#1",
		Name:         "Itlimoc, Cradle of the Sun",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{G}",
				Label:    "{T}: Add {G}",
			},
			{
				Cost:         ManaAbilityCost{Tap: true},
				ProducedFunc: ProducedPerPermanent("G", MatchCreature),
				Label:        "{T}: Add {G} for each creature you control",
			},
		},
	})
}
