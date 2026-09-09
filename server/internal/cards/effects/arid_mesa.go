package effects

// Arid Mesa — Land (EDHREC rank 57):
//
//	"{T}, Pay 1 life, Sacrifice this land: Search your library for
//	a Mountain or Plains card, put it onto the battlefield, then shuffle."
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
		OracleID: "c5acf2a5-40f4-433d-a74d-1cb56c521464",
		Name:     "Arid Mesa",
		Activated: []ActivatedAbility{{
			Label:  "{T}, Pay 1 life, Sacrifice this land: Search your library for a Mountain or Plains card, put it onto the battlefield, then shuffle.",
			Cost:   fetchlandCost(),
			Effect: fetchDual("mountain", "plains"),
		}},
	})
}
