package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Metamorphosis — Sorcery {G}:
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Add X mana of any one color, where X is 1 plus the sacrificed
//	 creature's mana value. Spend this mana only to cast creature
//	 spells."
//
// X is 1 plus the sacrificed creature's mana value as it last existed
// on the battlefield (CR 608.2h), read off the payment record (ADR 0113
// §1). The caster picks one colour and gets all X of it, each carrying
// the creature-spells-only restriction, which also lets it pay a
// creature spell's additional costs such as kicker (the 2004-10-04
// ruling).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:       "7140d726-0136-43af-84b5-85005a66a186",
		Name:           "Metamorphosis",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("a creature", Creature()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			x := 1 + sacrificedManaValue(ctx)
			ChooseColorTail(game.ColorForMana, ctx.Game, item.Controller, item.SourceCardID,
				"Metamorphosis — choose a color of mana",
				withN(ctx.ItemTail(metamorphosisColorChosen), x))
			return nil
		},
	})
}

// metamorphosisColorChosen adds the tail's N mana of the colour (its Word),
// spendable only on creature spells. An on-disk key.
var metamorphosisColorChosen = CardTail("then/metamorphosis-color-chosen", func(ctx *Context, o game.TailOutcome) error {
	return AddMana{
		Amount:       ManaN(o.Tail.N, o.Tail.Word),
		Restrictions: []string{ManaRestrictCast, ManaRestrictType("Creature")},
	}.Apply(ctx)
})
