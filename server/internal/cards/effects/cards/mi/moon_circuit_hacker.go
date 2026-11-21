package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Moon-Circuit Hacker — Enchantment Creature — Human Ninja {1}{U}, 2/1
// (Neon Dynasty):
//
//	"Ninjutsu {U}
//	 Whenever this creature deals combat damage to a player, you may draw a
//	 card. If you do, discard a card unless this creature entered this turn."
//
// The Celestus's linked "you may draw. If you do" clause: declining the draw
// declines the discard. "Entered this turn" is read as the answer is applied
// from the turn's entry tally, so a Hacker put in by ninjutsu this turn loots
// for free while one that entered on an earlier turn must discard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7cd8b017-8d59-4b55-bb9a-8d8e492ef1ac",
		Name:         "Moon-Circuit Hacker",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Ninjutsu("{U}")},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Moon-Circuit Hacker — you may draw a card; if you do, discard a card unless this creature entered this turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayChoice{
						Question: "Moon-Circuit Hacker — draw a card? (Then discard a card unless this creature entered this turn.)",
						Draws:    1,
						YesLabel: "Draw",
						NoLabel:  "Decline",
						Then:     moonCircuitHackerAnswer,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

var moonCircuitHackerAnswer = effects.MayYes("confirm/moon-circuit-hacker-draw", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
	if ctx.Game.EnteredThisTurn(ctx.Source()) {
		return ctx.Game.DrawNForEffect(ctx.Controller(), 1)
	}
	return effects.DrawThenDiscard(ctx.Game, ctx.Controller(), ctx.Source(), 1, 1, "")
})
