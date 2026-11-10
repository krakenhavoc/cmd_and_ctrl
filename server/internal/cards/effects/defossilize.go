package effects

// Defossilize — Sorcery {4}{B} (The Lost Caverns of Ixalan, #4470):
//
//	"Return target creature card from your graveyard to the battlefield.
//	 That creature explores, then it explores again."
//
// Three steps (CR 608.2c): the return, the first explore and the second,
// each waiting for the question the one before asked (the graveyard
// question of an explore that revealed a nonland card). The explorer is the
// permanent the card became: if it did not reach the battlefield (a
// replacement, or the target was illegal) nothing explores. A creature card
// that is not a creature once it is on the battlefield still explores (the
// ruling).
//
// No simplification.
func init() {
	explore := Instruction(func(ctx *Context) error {
		// The target left the graveyard as a card and is now a new
		// permanent, so it is read off the item, not re-checked as a
		// target (CR 400.7).
		if len(ctx.Item.Targets) == 0 || !ctx.Game.Battlefield.Contains(ctx.Item.Targets[0].ID) {
			return nil
		}
		return Explores{Explorer: ctx.Item.Targets[0].ID}.Apply(ctx)
	})
	Register(Spec{
		OracleID:     "d620de15-8141-4877-9330-8328a3fa85e1",
		Name:         "Defossilize",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("target creature card from your graveyard", Creature(), YouOwn()),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return returnTargetCardToTheBattlefield(ctx.Item, ctx) }),
			explore,
			explore,
		),
	})
}
