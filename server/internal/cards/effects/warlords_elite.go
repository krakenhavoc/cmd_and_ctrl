package effects

// Warlord's Elite — Creature — Human Soldier {2}{W} (#3868):
//
//	"As an additional cost to cast this spell, tap two untapped
//	 artifacts, creatures, and/or lands you control."
//
// The additional cost is New Blood's (TapUntappedCost) over any mix of
// the three types; a land tapped to pay it is not tapped for the {2}{W}.
// The body is the printed 4/4.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ee2c8714-6e85-4eb5-9bd6-057b6abfb136",
		Name:         "Warlord's Elite",
		Completeness: CompletenessFull,
		AdditionalCost: TapUntappedCost(2, "two untapped artifacts, creatures, and/or lands you control",
			Or(Artifact(), Creature(), Land())),
	})
}
