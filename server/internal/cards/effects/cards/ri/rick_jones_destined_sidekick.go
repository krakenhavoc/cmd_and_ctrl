package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rick Jones, Destined Sidekick — {G} Legendary Creature — Human Advisor,
// 0/3 (Marvel Super Heroes, #4455):
//
//	"{3}, {T}: Mill four cards. You may put a Hero or enchantment card
//	 from among those cards into your hand."
//
// Six's shape (mill_then_take.go): the choice is offered over the cards
// this mill put into the graveyard, re-checked against the graveyard so a
// card that has since moved is not offered. A Hero is read through the
// card's own text (a changeling counts), and an enchantment card is any
// enchantment, creature or not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4bccfa67-0c6f-42a1-b8d4-0842a8230d1e",
		Name:         "Rick Jones, Destined Sidekick",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: Mill four cards. You may put a Hero or enchantment card from among those cards into your hand.",
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.MillToZone{N: 4, ThenTail: ctx.ItemTail(rickJonesMilled)}.Apply(ctx)
			},
		}},
	})
}

// rickJonesMilled is the clause after the mill: a Hero or enchantment
// card from among the milled cards may go to the hand.
var rickJonesMilled = effects.CardTail("msh-a7/rick-jones-after-mill", func(ctx *effects.Context, o game.TailOutcome) error {
	return effects.MayTakeOneFromAmongThem(ctx, o.IDs, effects.Or(effects.Subtype("Hero"), effects.Enchantment()),
		"Rick Jones — you may put a Hero or enchantment card from among those cards into your hand")
})
