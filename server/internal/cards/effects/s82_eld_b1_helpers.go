package effects

// s82_eld_b1_helpers.go — shared bodies for the Throne of Eldraine cards
// of batch 1 (#5516). Append-only.

// eldB1IfAdamant is an adamant rider as a resolution step (CR 207.2c):
// `then` happens only when at least three `color` mana was spent to cast
// the spell. A waived payment reads as not spent (AdamantSpent), the
// weaker-than-printed direction. Unexplained Vision's scry 3.
func eldB1IfAdamant(color string, then Applier) Instruction {
	return func(ctx *Context) error {
		if !AdamantSpent(ctx, color, 3) {
			return nil
		}
		return then.Apply(ctx)
	}
}

// eldB1AdamantFood is "Adamant — If at least three <color> mana was spent
// to cast this spell, create a Food token." (Foreboding Fruit, Turn into
// a Pumpkin).
func eldB1AdamantFood(color string) Instruction {
	return eldB1IfAdamant(color, CreateToken{Template: FoodToken(), N: 1})
}
