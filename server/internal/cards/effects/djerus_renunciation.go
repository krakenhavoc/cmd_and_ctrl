package effects

// Djeru's Renunciation — Instant {1}{W} (Hour of Devastation, #5534):
//
//	"Tap up to two target creatures.
//	 Cycling {W} ({W}, Discard this card: Draw a card.)"
//
// Repel the Darkness's tap without its draw (the shared tapEachLegalTarget,
// which skips a target that left), and the shared Cycling row.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "11b21086-0779-4d16-8a56-252c6bee9bed",
		Name:         "Djeru's Renunciation",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve:    ResolveSteps(Instruction(func(ctx *Context) error { return tapEachLegalTarget(ctx.Game, ctx.Item) })),
		Activated:    []ActivatedAbility{Cycling("{W}")},
	})
}
