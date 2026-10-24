package effects

// Divide by Zero — Instant {2}{U} (EDHREC rank 17095):
//
//	"Return target spell or permanent with mana value 1 or greater to
//	 its owner's hand.
//	 Learn. (You may reveal a Lesson card you own from outside the game
//	 and put it into your hand, or discard a card to draw a card.)"
//
// One target that is either a spell on the stack or a permanent, with
// the mana value clause on both halves (X is zero on a permanent and the
// chosen value on a spell, CR 202.3e; a token with no cost is zero and
// is not a legal target). The return is Venser, Shaper Savant's, shared
// as ReturnTargetSpellOrPermanentToHand; learn follows it as a step
// (CR 701.48a, learn.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7b1c1a0c-f443-4eb5-a430-9eea886361ca",
		Name:         "Divide by Zero",
		Completeness: CompletenessFull,
		Targets: TargetSpellOrPermanent("target spell or permanent with mana value 1 or greater",
			ManaValueGE(1), ManaValueGE(1)),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return ReturnTargetSpellOrPermanentToHand(ctx.Game, ctx.Item)
			}),
			Learn{},
		),
	})
}
