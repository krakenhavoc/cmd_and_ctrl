package effects

// Grim Wanderer — Creature — Goblin Warlock {1}{B}, 5/3 (Adventures in the
// Forgotten Realms, S82):
//
//	"Flash
//	 Tragic Backstory — Cast this spell only if a creature died this turn."
//
// The timing clause is a CastCondition on the spell itself, read as it is
// announced, over the table-wide tally of creatures that died.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:           "614feadd-b74f-4d9c-a0a3-3a63724956af",
		Name:               "Grim Wanderer",
		Completeness:       CompletenessFull,
		PrintedKeywords:    []string{"flash"},
		CastCondition:      s82AfrB07CreatureDiedCast,
		CastConditionLabel: "Cast this spell only if a creature died this turn.",
	})
}
