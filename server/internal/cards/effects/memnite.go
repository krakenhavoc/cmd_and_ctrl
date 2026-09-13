package effects

// Memnite — Artifact Creature — Construct {0}, 1/1 (EDHREC rank
// 2287):
//
//	(no rules text)
//
// The free artifact body — the affinity deck's cheapest artifact and
// the Cathars' Crusade deck's cheapest trigger. Vanilla: a spec with
// no abilities, registered so the card is audited and published as
// complete rather than left with the "rules not implemented" chip a
// card the catalog does not know carries. The body, cost and type
// line are printed card data.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7663ac7c-1de3-4250-b96a-fae9dbd66a27",
		Name:         "Memnite",
		Completeness: CompletenessFull,
	})
}
