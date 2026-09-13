package effects

// Misty Rainforest — Land (EDHREC rank 42):
//
//	"{T}, Pay 1 life, Sacrifice this land: Search your library for
//	a Forest or Island card, put it onto the battlefield, then shuffle."
//
// One of the ten Zendikar / Onslaught fetchlands. The whole cycle is
// one card printed ten times with different subtypes, so the cost,
// the predicate and the search all live in fetchland_helpers.go and
// each file is just its own name, oracle ID and pair of types.
//
// The fetched land arrives UNTAPPED — that, plus the shuffle, is
// what the life is paying for. Evolving Wilds fetches tapped and
// costs no life, which is the entire difference between the two
// families.
func init() {
	Register(Spec{
		OracleID:     "09dd85aa-47bc-4713-a9b9-8b52ff2285ed",
		Name:         "Misty Rainforest",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:  "{T}, Pay 1 life, Sacrifice this land: Search your library for a Forest or Island card, put it onto the battlefield, then shuffle.",
			Cost:   fetchlandCost(),
			Effect: fetchDual("forest", "island"),
		}},
	})
}
