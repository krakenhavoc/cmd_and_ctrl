package effects

// Cursed Mirror — Artifact {2}{R}:
//
//	"{T}: Add {R}.
//	 As this artifact enters, you may have it become a copy of any
//	 creature on the battlefield until end of turn, except it has
//	 haste."
//
// S16.5's EntersAsCopyOf is a PERMANENT copy (Clone, Phyrexian
// Metamorph); Cursed Mirror wants a copy with a DURATION, which is
// the still-open half of the #665 row that also blocks Shifting
// Woodland's Delirium ability (docs/engine-seams.md, "become-copy-
// effects: Shifting Woodland" — Mirage Mirror, Cytoshape). Ships with
// only the mana ability until that primitive exists.
//
// Caveat: the ETB copy-until-end-of-turn isn't implemented — this
// artifact never becomes a copy of anything.
func init() {
	Register(Spec{
		OracleID:     "4d67e2a7-4aa7-44cc-853b-500d7aac046d",
		Name:         "Cursed Mirror",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Becoming a copy of a creature until end of turn isn't implemented — copy effects with a duration aren't built yet. Only the {T}: Add {R} ability works."},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "Add {R}",
		}},
	})
}
