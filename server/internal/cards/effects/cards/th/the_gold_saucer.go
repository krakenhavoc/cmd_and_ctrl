package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Gold Saucer — "{T}: Add {C}. {2}, {T}: Flip a coin. If you win the
// flip, create a Treasure token. {3}, {T}, Sacrifice two artifacts: Draw a
// card."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "93e38650-ce22-4ab9-b79d-cc7b6477c075",
		Name:         "The Gold Saucer",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "{T}: Add {C}",
		}},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}, {T}: Flip a coin. If you win the flip, create a Treasure token.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					g.FlipCoinForEffect(game.CoinFlipSpec{
						Flipper:  item.Controller,
						Source:   item.SourceCardID,
						Question: "The Gold Saucer — call the coin flip",
						ThenKey:  goldSaucerFlipped,
					})
					return nil
				},
			},
			{
				Label:   "{3}, {T}, Sacrifice two artifacts: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost(), effects.SacrificeN(2, "two artifacts", effects.Artifact())),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.DrawCards{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}

// goldSaucerFlipped is the settled call: a won flip makes a Treasure for
// the flipper, the ability's controller (#4253 slice 10b). An on-disk key:
// never renamed, never reused.
var goldSaucerFlipped = game.RegisterCoinFlipThen("coin-flip/the-gold-saucer-treasure", func(g *game.Game, r game.CoinFlipped) error {
	if effects.WonFirstFlip(r) {
		return effects.TreasureTokens(g, r.Flipper, 1)
	}
	return nil
})
