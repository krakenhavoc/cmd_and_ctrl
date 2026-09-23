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
// The mechanism is `DerivedMatch` — the produced string is computed at
// activation, after the cost is paid, from the opposing battlefield as
// it stands. "Could produce" is CR 106.7 and the card's own rulings:
// it asks what an opposing land's abilities WOULD add, not whether
// that land could legally be tapped right now. A tapped Island still
// offers {U}.
//
// Two things fall out of the wording and are worth stating:
//
//   - "any COLOR", so {C} is filtered out (DerivedColorsOnly — CR
//     105.1, colorless is not a colour). An opposing board of Wastes
//     and Ancient Tombs makes this land produce nothing.
//   - Nothing to derive from means nothing produced. Exotic Orchard
//     with no opposing lands taps for zero mana, which is what the
//     printed card does, and is why an empty derivation is a
//     supported answer rather than an error.
//
// An opposing land whose own ability is ITSELF derived — a second
// Exotic Orchard, a Reflecting Pool — is recursed into, not skipped
// (#1323): game.ProducibleManaLocked carries a visited set the whole
// way down, so a one-way chain (this Orchard → an opposing Reflecting
// Pool → THAT player's own plain Forest) resolves to a real colour.
// Simplification, declared and narrower than it used to be: a
// genuinely CIRCULAR chain — two Exotic Orchards facing each other, or
// an Orchard and a Reflecting Pool that both, directly or indirectly,
// end up asking about each other — still answers "no mana" for the
// permanents in the cycle (CR 106.6b), one colour short of what the
// real rules would resolve for some of those pairs. Weaker than
// printed.
//
// Auto-tap plans it fine: the derivation is a pure read, so the
// planner evaluates the same derivation the activation will.
func init() {
	Register(Spec{
		OracleID:     "27b047e3-0d41-45e2-98e9-9391d7923a1e",
		Name:         "Exotic Orchard",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Opposing lands that copy other lands' mana in a genuine circle back to this land (facing another Exotic Orchard, or a Reflecting Pool that in turn reads this land) contribute nothing."},
		ManaAbilities: []ManaAbility{{
			Cost:              ManaAbilityCost{Tap: true},
			DerivedMatch:      DerivedFromOpponentLands(),
			DerivedColorsOnly: true,
			// CR 106.6b: this ability reads what OTHER permanents
			// could produce, so CR 106.7's reader must route it
			// through the visited set rather than a plain ProducedFunc
			// call (#782, #1323).
			DerivesFromOtherSources: true,
			Label:                   "Add one mana of any color an opponent's land could produce",
			// The printed text says "any color that a land an
			// opponent controls could produce" — it does not mention
			// the commander's identity, so the derived option set
			// is not narrowed by it (NarrowToCommanderIdentity off).
		}},
	})
}
