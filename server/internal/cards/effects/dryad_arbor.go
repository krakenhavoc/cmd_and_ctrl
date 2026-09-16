package effects

// Dryad Arbor — Land Creature — Forest Dryad, 1/1 (EDHREC rank 762):
//
//	"(This land isn't a spell, it's affected by summoning sickness,
//	 and it has "{T}: Add {G}.")"
//
// The whole card is reminder text, and every clause of it is engine
// behaviour: it is played as a land because its type line says Land,
// it is summoning sick because its type line says Creature and the
// mana ability has a tap cost (CR 302.6, checked in
// ActivateManaAbility), and since #354 a Forest subtype produces {G}
// with no Spec at all. The ability is declared here anyway so the card
// reads like every other land in the catalog and its coverage is
// pinned.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e996cd67-739c-40f4-b276-0042acf26c71",
		Name:         "Dryad Arbor",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
	})
}
