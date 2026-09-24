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
// `DerivedMatch` is the filter.
//
// It also mattered more than the old note implied. An over-permissive
// simplification is the #259 direction: a Fellwar Stone that offers
// five colours off a mono-green table is strictly better than the
// printed card, not worse, and in a real game it fixed a colour no
// opponent could make. Of the batch-01 declared simplifications this
// was the only one pointing the wrong way.
//
// #1323 closed the OTHER declared gap: an opposing Exotic Orchard or
// Reflecting Pool used to contribute nothing at all, because the old
// recursion guard skipped every derived source unconditionally rather
// than only the ones that would actually cycle. Now
// game.ProducibleManaLocked carries the ancestor path, so a one-way
// chain (this Stone → an opposing Exotic Orchard → THAT player's own
// plain Forest) resolves to a real colour. An opposing land that,
// directly or indirectly, ends up asking about itself through this
// derivation with no real land anywhere in the loop still answers "no
// mana" — that is not a simplification, it is the printed card's own
// worked ruling (Exotic Orchard, word for word) and CR 106.7's own
// closing sentence; see game/producible_mana.go's file doc. Nothing
// here is weaker than printed.
func init() {
	Register(Spec{
		OracleID:     "95560508-7ac9-4be9-8a3f-3c7d5b52807b",
		Name:         "Fellwar Stone",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:              ManaAbilityCost{Tap: true},
			DerivedMatch:      DerivedFromOpponentLands(),
			DerivedColorsOnly: true,
			// CR 106.7: this ability reads what OTHER permanents
			// could produce, so CR 106.7's own reader must route it
			// through the ancestor-path guard rather than a plain
			// ProducedFunc call (#782, #1323).
			DerivesFromOtherSources: true,
			Label:                   "Add one mana of any color an opponent's land could produce",
			// The printed text derives the colours from opponents'
			// lands and says nothing about the commander. Before
			// #352 this card was one of the ones #276 called out as
			// wrongly narrowed by commander identity; now the
			// derivation IS the narrowing, and
			// NarrowToCommanderIdentity stays off.
		}},
	})
}
