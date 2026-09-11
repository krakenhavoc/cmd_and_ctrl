package effects

// signets.go — the ten Ravnica Signets:
//
//	Artifact {2}
//	"{1}, {T}: Add {A}{B}."
//
// Ten cards that look trivial and were, until #352, literally
// unwritable: `ManaAbilityCost` carried Tap, Sacrifice, SacrificeOther
// and Life, and had no mana component at all. The card-coverage
// roadmap singles the cycle out as the cheapest sub-gap in the whole
// rank-2 mana pipeline, and it was — one field on the cost struct, one
// validate-and-pay step in ActivateManaAbility.
//
// The whole cycle registers together because a Signet is data: the
// shape is stated once here and each card is three strings. Ranks run
// 159 (Dimir) to 1075 (Selesnya), so every one of them is inside the
// roadmap's top 2000.
//
// What the {1} means mechanically:
//
//   - It is a COST, validated before anything else is paid, so a
//     Signet activated on an empty pool fails with the artifact still
//     untapped. (Contrast a rider, which happens regardless — see
//     Ancient Tomb.)
//   - There is no auto-tap into it. ActivateManaAbility deliberately
//     will not tap other permanents to fund a mana ability: a mana
//     ability resolves with no priority window (CR 605.3a), and the
//     planner cannot weigh "tap three lands to filter one". The
//     player floats the {1} and clicks the Signet, which is exactly
//     how the card is played on paper.
//   - Consequently the auto-tapper does not plan Signets as sources
//     either (autoTapAbilityFor skips any ability with a mana cost).
//     A Signet already tapped for mana pays for the next spell like
//     any other mana in the pool; it just is not tapped FOR you.
//
// Net-positive filtering: {1} in, {A}{B} out, so a Signet ramps by
// one and fixes by two. That is the whole card and it is all here.
//
// No simplification. IgnoreCommanderIdentity is not set and does not
// need to be — the produced string names two specific colours as two
// separate slots, not a pipe, so there is no option set to narrow.
func init() {
	for _, s := range []struct{ oracleID, name, a, b string }{
		{"7d881c57-0bd9-4c57-aa4a-b10808b86143", "Dimir Signet", "U", "B"},
		{"3adb7681-977f-4a32-9ec8-51481b958268", "Rakdos Signet", "B", "R"},
		{"2fda4fe7-8b0c-489c-a000-6d358e614e34", "Izzet Signet", "U", "R"},
		{"de3dcb5d-775a-479f-99f5-d1883ed9b1b5", "Orzhov Signet", "W", "B"},
		{"e018773f-95b3-49a3-9674-6f04ddef2092", "Azorius Signet", "W", "U"},
		{"41c84665-1f99-40ab-aaca-1188649eb263", "Boros Signet", "R", "W"},
		{"44503105-3e13-408d-a44f-37d503c61d72", "Simic Signet", "G", "U"},
		{"1cf51f50-24e4-48d0-95b3-1dad3ffa4bf5", "Golgari Signet", "B", "G"},
		{"d36e0c9f-c025-4dfe-9644-9cad2461ce38", "Gruul Signet", "R", "G"},
		{"1436dd81-496e-42a5-b210-fb5b9cdf073f", "Selesnya Signet", "G", "W"},
	} {
		produced := "{" + s.a + "}{" + s.b + "}"
		Register(Spec{
			OracleID:     s.oracleID,
			Name:         s.name,
			Completeness: CompletenessFull,
			ManaAbilities: []ManaAbility{{
				Cost:     ManaAbilityCost{Tap: true, Mana: "{1}"},
				Produced: produced,
				Label:    "{1}, {T}: Add " + produced,
			}},
		})
	}
}
