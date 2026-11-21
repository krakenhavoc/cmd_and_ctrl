package effects

// Repudiate // Replicate — split card {G/U}{G/U} // {1}{G}{U} (Ravnica
// Remastered):
//
//	Repudiate (Instant): "Counter target activated or triggered ability.
//	 (Mana abilities can't be targeted.)"
//	Replicate (Sorcery): "Create a token that's a copy of target creature
//	 you control."
//
// Stifle's counter and Cackling Counterpart's token copy. The left half is
// the bare oracle ID and the right half "<oracle>#1" (ADR 0034 faces).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a43ecf8b-5dc9-4ac4-ae1f-b663aa8c9d7b",
		Name:         "Repudiate",
		Completeness: CompletenessFull,
		Targets:      AbilityOnStack("target activated or triggered ability"),
		OnResolve:    counterTheTargetSpell,
	})
	Register(Spec{
		OracleID:     "a43ecf8b-5dc9-4ac4-ae1f-b663aa8c9d7b#1",
		Name:         "Replicate",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you control", YouControl()),
		OnResolve:    TokenCopyOfSingleTarget,
	})
}
