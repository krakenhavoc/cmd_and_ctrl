package effects

// Exotic Orchard — Land:
//
//	"{T}: Add one mana of any color that a land an opponent controls
//	 could produce."
//
// Rank 9. The highest-ranked card the catalog did not have, in the
// entire format — and the card the coverage roadmap named as the
// headline of the rank-2 mana pipeline. Until #352 it was unwritable:
// `Produced` was a static string, and nothing in the engine could
// compute one from the board.
//
// The mechanism is `ProducedFunc` — the produced string is computed at
// activation, after the cost is paid, from the opposing battlefield as
// it stands. "Could produce" is CR 106.7 and the card's own rulings:
// it asks what an opposing land's abilities WOULD add, not whether
// that land could legally be tapped right now. A tapped Island still
// offers {U}.
//
// Two things fall out of the wording and are worth stating:
//
//   - "any COLOR", so {C} is filtered out (CR 105.1 — colorless is not
//     a colour). An opposing board of Wastes and Ancient Tombs makes
//     this land produce nothing.
//   - Nothing to derive from means nothing produced. Exotic Orchard
//     with no opposing lands taps for zero mana, which is what the
//     printed card does, and is why ProducedFunc returning "" is a
//     supported answer rather than an error.
//
// Simplification, declared: an opposing land whose own ability is
// itself derived — a second Exotic Orchard, a Reflecting Pool — is
// skipped rather than recursed into (see mana_derivation.go's
// recursion guard). CR 106.6b answers the genuinely circular case with
// "no mana" and this agrees; where the real rules would resolve a
// one-way chain, this is one colour short. Weaker than printed.
//
// Auto-tap plans it fine: the derivation is a pure read, so the
// planner evaluates the same ProducedFunc the activation will.
func init() {
	Register(Spec{
		OracleID: "27b047e3-0d41-45e2-98e9-9391d7923a1e",
		Name:     "Exotic Orchard",
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true},
			ProducedFunc: ProducedFromOpponentLands(),
			Label:        "Add one mana of any color an opponent's land could produce",
			// The printed text says "any color that a land an
			// opponent controls could produce" — it does not mention
			// the commander's identity, so the derived option set
			// must not be narrowed by it.
			IgnoreCommanderIdentity: true,
		}},
	})
}
