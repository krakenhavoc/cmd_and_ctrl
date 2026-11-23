package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inquisitor Eisenhorn — Legendary Creature — Human Inquisitor {2}{U}{B}, 2/3:
//
//	"You may reveal the first card you draw each turn as you draw it. Whenever
//	 you reveal an instant or sorcery card this way, create Cherubael, a
//	 legendary 4/4 black Demon creature token with flying.
//	 Whenever Inquisitor Eisenhorn deals combat damage to a player, investigate
//	 that many times."
//
// The reveal and its trigger are one step here: when the first card you draw
// each turn is an instant or sorcery, you are asked whether to reveal it, and
// a yes creates Cherubael. A first card that is anything else has no prompt,
// since revealing it does nothing and only shows it to the table.
var inquisitorEisenhornReveal = effects.MayYes("confirm/inquisitor-eisenhorn-reveal", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
	return effects.CreateToken{Template: effects.TokenCard("4/4 black Legendary Demon with flying named Cherubael"), N: 1}.Apply(ctx)
})

func init() {
	effects.Register(effects.Spec{
		OracleID:     "db07a3f9-c6d1-4cae-a899-12b3e8d94018",
		Name:         "Inquisitor Eisenhorn",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller || g.TurnTallyFor(source.Controller).CardsDrawn != 1 {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && (c.IsInstant() || c.HasCardType("sorcery"))
			}, "Inquisitor Eisenhorn — you may reveal the first card you draw each turn", func(g *game.Game, item *game.StackItem) error {
				return effects.MayChoice{
					Question: "Inquisitor Eisenhorn — reveal the instant or sorcery card you drew to create Cherubael?",
					Then:     inquisitorEisenhornReveal,
				}.Apply(effects.NewContext(g, item))
			}),
			{
				Watches:   []game.EventKind{game.EventDealDamage},
				AppliesTo: effects.ThisDealtCombatDamageToAPlayer,
				Key:       "Inquisitor Eisenhorn — investigate that many times",
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, "Inquisitor Eisenhorn — investigate that many times")
					item.Params.Amount = ev.Amount
					return item
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.Investigate{N: item.Params.Amount}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
