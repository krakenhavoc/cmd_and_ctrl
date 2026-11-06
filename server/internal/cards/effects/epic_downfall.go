package effects

// Epic Downfall — Sorcery {1}{B}:
//
//	"Exile target creature with mana value 3 or greater."
//
// Death in the Family's shape with the bound the other way. The mana
// value is the permanent's on the battlefield, where an {X} in its cost
// is 0 (CR 202.3e, the ruling), and it is checked as the spell is cast
// and again as it resolves (CR 608.2b).
//
// No simplification. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:     "ed5d6aba-0b8d-48b3-a84c-618f45fb67b1",
		Name:         "Epic Downfall",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature with mana value 3 or greater", ManaValueGE(3)),
		OnResolve:    ResolveSteps(Instruction(exileTheFirstTargetIfLegal)),
	})
}
