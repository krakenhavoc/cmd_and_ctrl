package effects

// Graven Cairns — Land (EDHREC rank 555):
//
//	"{T}: Add {C}.
//	 {B/R}, {T}: Add {B}{B}, {B}{R}, or {R}{R}."
//
// The Shadowmoor filter land: one mana of either colour in, two mana
// in any mix of the two out. Unwritable until #356 put a mana
// component on ManaAbilityCost — the same seam the Signets and the
// Odyssey filter lands waited on — and now three strings:
//
//   - The cost is the hybrid symbol as printed. ParseCost reads
//     {B/R} as one requirement payable by either colour, and the
//     engine validates it against the pool before the land taps, so
//     an activation with only {G} floating fails with the Cairns
//     untouched. There is no auto-tap into it (see signets.go); the
//     player floats the mana and clicks.
//   - The output is two independent {B|R} slots. "BB, BR, or RR" is
//     exactly the set of two independent picks from {B, R}, so two
//     colour prompts give the printed choice space with no third
//     shape. NarrowToCommanderIdentity stays off: the card names
//     its colours and says nothing about the command zone.
//
// The colorless half sits at index 0 so the auto-tapper reaches for
// it and never spends the player's floating mana on a filter.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5004b84a-33b7-4f6f-b2c2-7086b9087535",
		Name:         "Graven Cairns",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			painlessColorless(),
			{
				Cost:     ManaAbilityCost{Tap: true, Mana: "{B/R}"},
				Produced: "{B|R}{B|R}",
				Label:    "{B/R}, {T}: Add {B}{B}, {B}{R}, or {R}{R}",
			},
		},
	})
}
