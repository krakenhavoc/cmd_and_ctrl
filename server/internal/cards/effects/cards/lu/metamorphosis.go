package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:       "7140d726-0136-43af-84b5-85005a66a186",
		Name:           "Metamorphosis",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeCost("a creature", effects.Creature()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			x := 1 + effects.SacrificedManaValue(ctx)
			effects.ChooseColorTail(game.ColorForMana, ctx.Game, item.Controller, item.SourceCardID,
				"Metamorphosis — choose a color of mana",
				effects.WithN(ctx.ItemTail(metamorphosisColorChosen), x))
			return nil
		},
	})
}

// metamorphosisColorChosen adds the tail's N mana of the colour (its Word),
// spendable only on creature spells. An on-disk key.
var metamorphosisColorChosen = effects.CardTail("then/metamorphosis-color-chosen", func(ctx *effects.Context, o game.TailOutcome) error {
	return effects.AddMana{
		Amount:       effects.ManaN(o.Tail.N, o.Tail.Word),
		Restrictions: []string{effects.ManaRestrictCast, effects.ManaRestrictType("Creature")},
	}.Apply(ctx)
})
