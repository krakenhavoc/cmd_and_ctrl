package effects

// Elvish Aberration — Creature — Elf Mutant {5}{G}, 4/5 (Scourge):
//
//	"{T}: Add {G}{G}{G}.
//	 Forestcycling {2} ({2}, Discard this card: Search your library for a
//	 Forest card, reveal it, put it into your hand, then shuffle.)"
//
// Canopy Tactician's {T}: Add {G}{G}{G} mana ability (CR 605.1a), which a
// creature can't use the turn it arrives unless it has haste, and the
// shared typecycling ability (CR 702.29e).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1d25b831-c1de-4f79-bbb3-1582eda96452",
		Name:         "Elvish Aberration",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}{G}{G}",
			Label:    "Add {G}{G}{G}",
		}},
		Activated: []ActivatedAbility{
			Typecycling("Forestcycling", "{2}", "a Forest card", IsLandWithSubtype("Forest")),
		},
	})
}
