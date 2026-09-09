package effects

// Flooded Strand — Land (EDHREC rank 39):
//
//	"{T}, Pay 1 life, Sacrifice this land: Search your library for
//	a Plains or Island card, put it onto the battlefield, then shuffle."
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
		OracleID: "f3c7af78-a77d-4134-82a2-a5ce84285a84",
		Name:     "Flooded Strand",
		Activated: []ActivatedAbility{{
			Label:  "{T}, Pay 1 life, Sacrifice this land: Search your library for a Plains or Island card, put it onto the battlefield, then shuffle.",
			Cost:   fetchlandCost(),
			Effect: fetchDual("plains", "island"),
		}},
	})
}
