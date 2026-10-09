package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Soul Burn — Sorcery {X}{2}{B}:
//
//	"Spend only black and/or red mana on X.
//	 Soul Burn deals X damage to any target. You gain life equal to the
//	 damage dealt, but not more than the amount of {B} spent on X, the
//	 player's life total before the damage was dealt, the planeswalker's
//	 loyalty before the damage was dealt, or the creature's toughness."
//
// #2556: the "black and/or red" clause is SpellSpendOnlyOnX("B", "R").
//
// Simplification: the payment record knows how much black mana paid for
// the WHOLE spell, not which of it went to X, and {2}{B} can absorb up
// to (total - X) of it. So the black "spent on X" is counted as the
// least it could have been: black spent, minus every mana that paid
// something other than X. That is never more than printed, and exact
// when the {2}{B} was paid with other mana; the caveat says so.
func init() {
	Register(Spec{
		OracleID:     "063b0f5d-af27-4681-87b9-b553f6887061",
		Name:         "Soul Burn",
		XMatters:     true,
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The life you gain counts black mana spent on X as the least it could have been, because the game can't tell which black mana paid for X rather than for the rest of the cost, so you may gain less life than printed.",
		},
		SpendOnly: SpellSpendOnlyOnX("B", "R"),
		Targets:   TargetAny(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return damageThenGainUpToWhatItWas(ctx, item, blackSpentOnXAtLeast(ctx))
		},
	})
}

// blackSpentOnXAtLeast is the fewest {B} that can have paid for X: all
// the black mana spent, less every mana that paid for something other
// than X (the spell's {2}{B} and anything a cost modifier added). Zero
// for a payment the engine did not record.
//
// The mana that paid for X is X, unless a cost reduction ate into it
// (#2701): a generic reduction takes the {2} first and then the X, so
// once it reaches the X only the {B} is left beside it, and the X mana
// paid is the total less that one symbol. No cost modifier adds a
// coloured symbol to Soul Burn (only a spell's own strive-style
// increase does), so the {B} is the whole of the coloured part.
func blackSpentOnXAtLeast(ctx *Context) int {
	spent := ctx.ManaSpent()
	if !spent.Known() {
		return 0
	}
	const colouredSymbols = 1 // Soul Burn's {B}
	xPaid := max(0, min(ctx.X(), spent.Total()-colouredSymbols))
	elsewhere := spent.Total() - xPaid
	return max(0, min(xPaid, ctx.ManaSpentOfColor("B")-max(0, elsewhere)))
}
