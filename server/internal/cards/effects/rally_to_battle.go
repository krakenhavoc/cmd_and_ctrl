package effects

// Rally to Battle — Instant {3}{W} (Ravnica: Clue Edition, #5929):
//
//	"Creatures you control get +1/+3 until end of turn. Untap them."
//
// The boost is fixed on the creatures you control as the spell resolves
// (CR 611.2c) and "them" is that same set, untapped right after.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9d8d88dd-c266-4b7d-a756-ac2d8976ca89",
		Name:         "Rally to Battle",
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 1, Toughness: 3, Label: "Rally to Battle — +1/+3"}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error { return UntapAllCreaturesYouControl{}.Apply(ctx) }),
		),
	})
}
