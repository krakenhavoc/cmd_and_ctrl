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
// #3941 (ADR 0068, amendment 2026-10-16): "the amount of {B} spent on X"
// is a recorded fact. The pool solver keeps, for each mana it spends,
// whether it paid the restricted {X} part of the cost (ManaToken.ForX),
// and this card reads ManaSpent().OnX().Count("B"). That is exact for a
// payment the engine made, and for a waived one whose caster declared
// what they paid (the picker marks the mana that paid X).
//
// Which mana pays X is the caster's choice (CR 601.2h). The auto-pay
// solver makes it for them, and the card declares WantsColors B so it
// makes it in the card's favour: the printed {B} is paid first, and the
// X part then takes black before red, so the life gained is the most the
// mana in the pool allows. A payment declared on paper states its own
// allocation, and the record keeps it.
func init() {
	Register(Spec{
		OracleID:       "063b0f5d-af27-4681-87b9-b553f6887061",
		Name:           "Soul Burn",
		ReadsManaSpent: game.ReadsOwnPayment,
		WantsColors:    []string{"B"},
		XMatters:       true,
		Completeness:   CompletenessFull,
		SpendOnly:      SpellSpendOnlyOnX("B", "R"),
		Targets:        TargetAny(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return damageThenGainUpToWhatItWas(ctx, item, blackSpentOnX(ctx))
		},
	})
}

// blackSpentOnX is "the amount of {B} spent on X": the black mana the
// payment recorded against the {X} part of the cost. Zero for a payment
// the engine did not record (a copy, a payment waived with nothing
// declared), which is the weaker-than-printed answer.
func blackSpentOnX(ctx *Context) int {
	return ctx.ManaSpent().OnX().Count("B")
}
