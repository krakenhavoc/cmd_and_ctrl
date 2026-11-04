package effects

// Death in the Family — {1}{B} Instant (Teenage Mutant Ninja Turtles, #4457):
//
//	"Exile target creature with mana value 3 or less."
//
// Cruel Alliance's target clause without the teamwork. The mana value is
// checked again as the spell resolves (CR 608.2b), so a creature that grew
// past 3 in response is not exiled.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f599471a-31a6-4fb0-bf27-c5b707e1db6a",
		Name:         "Death in the Family",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature with mana value 3 or less", ManaValueLE(3)),
		OnResolve:    ResolveSteps(Instruction(exileTheFirstTargetIfLegal)),
	})
}
