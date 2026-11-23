package cardsgr

import (
	"errors"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grandmother Goby — Legendary Creature — Merfolk Wizard {2}{U}{U}, 2/2
// (Mystery Booster Commander Edition):
//
//	"Whenever you cast a noncreature spell, counter that spell. Create a
//	 number of 1/1 blue Merfolk creature tokens equal to the amount of mana
//	 spent to cast that spell."
//
// The amount is the mana actually spent, read off the spell's payment
// record as it is cast (Aberrant Manawurm's Build) and carried on the item,
// so it is known after the counter has taken the spell off the stack. The
// counter happens first and the tokens are made whether or not it worked:
// a spell that can't be countered still pays out (the tokens are not a
// "then" of the counter). A spell cast without paying its mana cost spent
// nothing: no tokens.
//
// No simplification.
func init() {
	const label = "Grandmother Goby — counter that spell, then create a 1/1 Merfolk for each mana spent to cast it"
	t := effects.WheneverYouCast(effects.Noncreature(), label, grandmotherGobyCounter)
	t.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Amount = g.StackItemPaidForEffect(ev.CardID).Spent().Total()
		item.Params.Object = game.ObjectRef{ID: ev.CardID, Epoch: g.ObjectEpochForEffect(ev.CardID)}
		return item
	}
	effects.Register(effects.Spec{
		OracleID:       "2d29c40a-4e13-4d99-bb5a-e991e385b8c2",
		Name:           "Grandmother Goby",
		Completeness:   effects.CompletenessFull,
		ReadsManaSpent: game.ReadsYourCasts,
		Triggered:      []game.TriggeredAbility{t},
	})
}

func grandmotherGobyCounter(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if err := g.CounterTargetForEffect(item.Params.Object.ID); err != nil && !errors.Is(err, game.ErrCardNotOnStack) {
		return err
	}
	if n := item.Params.Amount; n > 0 {
		return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("1/1 blue Merfolk"), N: n}.Apply(ctx)
	}
	return nil
}
