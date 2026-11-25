package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Niambi, Esteemed Speaker — Legendary Creature — Human Cleric {W}{U}, 2/1
// (Core Set 2021, #5511):
//
//	"Flash
//	 When Niambi enters, you may return another target creature you control to
//	 its owner's hand. If you do, you gain life equal to that creature's mana
//	 value.
//	 {1}{W}{U}, {T}, Discard a legendary card: Draw two cards."
//
// The trigger is Jeskai Barricade's optional targeted bounce. The mana value
// is read before the creature leaves (CR 608.2h) and carried to the bounce's
// continuation, which gains the life only if the card really reached a hand
// ("if you do"; a commander sent to the command zone instead gains nothing).
// The discard is a cost: only a legendary card in hand can pay it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "91c9dbd8-a8d8-4a4e-8408-3ad92e974d11",
		Name:            "Niambi, Esteemed Speaker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.Optional(
					effects.WhenThisEnters("Niambi, Esteemed Speaker — return another target creature you control to its owner's hand; if you do, you gain life equal to its mana value", niambiBounce),
					"Niambi, Esteemed Speaker — return another creature you control to its owner's hand?"),
				effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}{U}, {T}, Discard a legendary card: Draw two cards.",
			Purpose: game.Purpose{Draws: 2, Answers: game.AnswerValue},
			Cost: effects.Plus(effects.ManaCost("{1}{W}{U}"), effects.TapCost(),
				effects.DiscardCardsMatching(1, "a legendary card", func(c game.Card) bool { return effects.IsLegendary(&c) })),
			Effect: effects.Do(effects.DrawCards{N: 2}),
		}},
	})
}

// niambiBounce returns the announced creature and, behind the move, gains
// life equal to the mana value it had.
func niambiBounce(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok || !effects.OnBattlefield(g, id) {
		return nil
	}
	c, ok := effects.BattlefieldCardForEffect(g, id)
	if !ok {
		return nil
	}
	mv, _ := g.ManaValueForEffect(c)
	return effects.BounceToHand{Target: id, ThenTail: ctx.TailN(niambiGainLife, mv)}.Apply(ctx)
}

// niambiGainLife is "if you do, you gain life equal to that creature's mana
// value", Tail.N of it. An on-disk key.
var niambiGainLife = effects.CardTail("then/s82-old-b105-niambi-gain-life", func(ctx *effects.Context, o game.TailOutcome) error {
	if !effects.OneWentThisWay(o) {
		return nil
	}
	return effects.GainLife{Player: o.Tail.Player, Amount: o.Tail.N}.Apply(ctx)
})
