package effects

// Machine God's Effigy — Artifact {4}:
//
//	"You may have this artifact enter as a copy of any creature on
//	 the battlefield, except it's an artifact and it has '{T}: Add
//	 {U}.' (It's not a creature.)
//	 {T}: Add {U}."
//
// The copy clause is left out on purpose. CR 707.9b's "except it's an
// artifact" only ADDS the Artifact type (game.PrintedValues has
// AddCardType, no RemoveCardType); the parenthetical "(It's not a
// creature.)" needs the copied creature's OWN type stripped, which is
// the exact per-instance-type-override gap Enduring Tenacity and
// Enduring Curiosity already declare — the layer-4 machinery is keyed
// on the catalog entry a copy shares with every other printing.
// Shipping the copy without stripping Creature would leave Machine
// God's Effigy able to attack, block, and be affected by every
// creature-matters card as whatever it copied — strictly STRONGER
// than printed, which #259 forbids. So the card registers only its
// unconditional half: an artifact that taps for {U}, exactly what it
// is if the copy is always declined.
func init() {
	Register(Spec{
		OracleID:     "64ebdd6f-acde-4aab-a86b-2798bad5f70c",
		Name:         "Machine God's Effigy",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"This never enters as a copy of a creature — it's always just a plain artifact that taps for {U}.",
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "Add {U}",
		}},
	})
}
