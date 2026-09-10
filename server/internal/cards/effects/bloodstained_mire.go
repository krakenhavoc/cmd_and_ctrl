package effects

// Bloodstained Mire — Land (EDHREC rank 43):
//
//	"{T}, Pay 1 life, Sacrifice this land: Search your library for
//	a Swamp or Mountain card, put it onto the battlefield, then shuffle."
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
		OracleID: "fc0707c7-d504-4ccf-a0d2-3eb6e26e7a57",
		Name:     "Bloodstained Mire",
		Activated: []ActivatedAbility{{
			Label:  "{T}, Pay 1 life, Sacrifice this land: Search your library for a Swamp or Mountain card, put it onto the battlefield, then shuffle.",
			Cost:   fetchlandCost(),
			Effect: fetchDual("swamp", "mountain"),
		}},
	})
}
