package effects

// Phyrexian Walker — Artifact Creature — Phyrexian Construct {0},
// 0/3 (EDHREC rank 2963). No rules text.
//
// A vanilla 0/3 for nothing. The catalog entry exists so the card
// loses the "rules not implemented" chip: there are no rules to
// implement, and the {0} cost and the 0/3 body are printed card data
// the deck importer already carries.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7af75024-6c9b-4844-aeb6-81de25464822",
		Name:         "Phyrexian Walker",
		Completeness: CompletenessFull,
	})
}
