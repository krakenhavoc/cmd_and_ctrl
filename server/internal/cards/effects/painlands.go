package effects

// painlands.go — the eight most-played "painlands" (Ice Age's allied
// five, Apocalypse's enemy five; the two the top-100 triage did not
// rank are left for whoever wants the full ten):
//
//	"{T}: Add {C}."
//	"{T}: Add {A} or {B}. This land deals 1 damage to you."
//
// Two separate abilities, not one ability with an optional rider.
// The colorless half is genuinely free; the colored half always
// hurts. That distinction is the card, and it is why the auto-tapper
// can use a painland safely (see painlessColorless) while a human
// clicking the colored line knows what they signed up for.
//
// The damage is a RIDER, not a cost: "This land deals 1 damage to
// you" sits after the "Add" clause in the same ability, so it happens
// on activation with no stack and no way to decline, and the land
// stays activatable at 1 life. Compare Mana Confluence, whose "Pay 1
// life" IS a cost and is checked before the land taps.
//
// Grouped in one file for the reason temples.go states: eight
// near-identical cards in eight files is eight places to fix the same
// mistake. Each card is one row of data.
//
// No simplification. Both abilities, the damage, the damage's source
// and its interaction with the replacement pipeline are all as
// printed.
func init() {
	for _, land := range []struct{ oracleID, name, a, b string }{
		// Apocalypse (enemy pairs).
		{"0fe16212-66c3-4e45-a641-7391e9b2e304", "Shivan Reef", "U", "R"},
		{"6b75b94e-83b7-457e-ac41-7ca90b5a59aa", "Battlefield Forge", "R", "W"},
		{"33de01e9-ce5a-42d4-afcb-343cd54a6d80", "Caves of Koilos", "W", "B"},
		{"40b36bc6-c185-4bda-99e7-0118953c2c97", "Yavimaya Coast", "G", "U"},
		{"32116127-cf96-4a1b-8896-a1ebc087b597", "Llanowar Wastes", "B", "G"},
		// Ice Age / Apocalypse allied pairs.
		{"857febd9-cdd7-4f8e-a852-d88084b0cfbc", "Underground River", "U", "B"},
		{"d5ad26cc-2bdb-46b7-b8bf-dd099d5fa09b", "Adarkar Wastes", "W", "U"},
		{"f5c38c01-4a40-469f-91a0-7479daf4e8e7", "Sulfurous Springs", "B", "R"},
	} {
		Register(Spec{
			OracleID: land.oracleID,
			Name:     land.name,
			ManaAbilities: []ManaAbility{
				painlessColorless(),
				painDual(land.a, land.b, "This land"),
			},
		})
	}
}
