package effects

// Ethereal Guidance — Sorcery {2}{W} (Shadows over Innistrad, #5543):
//
//	"Creatures you control get +2/+1 until end of turn."
//
// On the Job's +2/+1 to the creatures you control until end of turn,
// without the Clue.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0e24c0d2-2333-425c-952e-9a13e6519cd3",
		Name:         "Ethereal Guidance",
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 2, Toughness: 1, Label: "Ethereal Guidance — +2/+1"},
		),
	})
}
