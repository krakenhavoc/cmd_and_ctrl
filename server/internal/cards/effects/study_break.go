package effects

// Study Break — Instant {1}{W} (EDHREC rank 25155):
//
//	"Tap up to two target creatures.
//	 Learn. (You may reveal a Lesson card you own from outside the game
//	 and put it into your hand, or discard a card to draw a card.)"
//
// "Up to two" lets it be cast with no creature at all, for the learn; a
// creature that left in response is skipped (CR 608.2b) and the others
// still tap. Learn is the last instruction (CR 701.48a, learn.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7aae7c7b-0b31-4b48-a8d3-81c913f604cf",
		Name:         "Study Break",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				for _, t := range ctx.LegalTargets() {
					if err := (TapTarget{Target: t.ID}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
			Learn{},
		),
	})
}
