package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Surge Mare — Creature — Horse Fish {U}{U}, 0/5 (Core Set 2019, #5527):
//
//	"This creature can't be blocked by green creatures.
//	 Whenever this creature deals damage to an opponent, you may draw a
//	 card. If you do, discard a card.
//	 {1}{U}: This creature gets +2/-2 until end of turn."
//
// Sootwalkers' block rule with green, Vedalken Heretic's "deals damage to
// an opponent" condition (any damage, not only combat damage) with Sage of
// the Falls' optional loot, and Hackrobat's pump.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8342de86-5b11-486d-8012-f0f9d2ded8c4",
		Name:         "Surge Mare",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.And(effects.Creature(), effects.OfColor("G")), "green creatures"),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.SnakeUmbraDealtDamageToOpponent,
				"Surge Mare — you may draw a card; if you do, discard a card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayChoice{
						Question: "Surge Mare — draw a card, then discard a card?",
						Draws:    1,
						YesLabel: "Draw",
						NoLabel:  "Decline",
						Then:     effects.CelestusAnswer,
					}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}: This creature gets +2/-2 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{1}{U}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(2, -2, "Surge Mare — +2/-2 until end of turn"),
		}},
	})
}
