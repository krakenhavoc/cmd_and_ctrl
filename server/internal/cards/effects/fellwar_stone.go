package effects

// Fellwar Stone — Artifact {2}:
//
//	"{T}: Add one mana of any color that a land an opponent controls
//	 could produce."
//
// Exotic Orchard on a two-mana rock, word for word, and rank 17 in
// its own right.
//
// #352 fixed this card rather than adding it. It has been registered
// since the top-100 batch with a declared simplification: the produced
// set was the full five-colour pipe instead of the opponents'
// intersection, because "narrowing needs a per-activation scan of
// every opponent's battlefield lands and their mana abilities; the
// engine has an equivalent hook only for commander colour identity".
// That note ended "Follow-up work is a ManaAbility-level filter
// analogous to commanderIdentityFor" — this is that follow-up, and
// `ProducedFunc` is the filter.
//
// It also mattered more than the old note implied. An over-permissive
// simplification is the #259 direction: a Fellwar Stone that offers
// five colours off a mono-green table is strictly better than the
// printed card, not worse, and in a real game it fixed a colour no
// opponent could make. Of the batch-01 declared simplifications this
// was the only one pointing the wrong way.
//
// The remaining simplification is the shared recursion guard
// documented in mana_derivation.go: an opposing land whose own output
// is derived contributes nothing. Weaker than printed.
func init() {
	Register(Spec{
		OracleID:     "95560508-7ac9-4be9-8a3f-3c7d5b52807b",
		Name:         "Fellwar Stone",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"It can't see colors from opponents' lands that themselves copy other lands' mana (Exotic Orchard, Reflecting Pool) — those contribute nothing."},
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true},
			ProducedFunc: ProducedFromOpponentLands(),
			Label:        "Add one mana of any color an opponent's land could produce",
			// The printed text derives the colours from opponents'
			// lands and says nothing about the commander. Before
			// #352 this card was one of the ones #276 called out as
			// wrongly narrowed by commander identity; now the
			// derivation IS the narrowing.
			IgnoreCommanderIdentity: true,
		}},
	})
}
