package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nadier, Agent of the Duskenel — Legendary Creature — Elf Warrior {5}{B},
// 3/3 (Commander Legends, #5507):
//
//	"Whenever a token you control leaves the battlefield, put a +1/+1 counter
//	 on Nadier.
//	 When Nadier leaves the battlefield, create a number of 1/1 green Elf
//	 Warrior creature tokens equal to its power.
//	 Partner"
//
// The first ability is Nadier's Nightblade's token departure (the token's
// controller as it last existed, CR 603.10a), including Nadier's own tokens
// and a token copy of itself. The second reads Nadier's power as it last
// existed on the battlefield (CR 608.2h), counters included; no tokens are made
// for a power of zero or less. Partner is a deck-construction rule.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "42b12ba0-7d1d-48f9-9473-881f7a19514a",
		Name:         "Nadier, Agent of the Duskenel",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && effects.IsToken(c) && effects.LeftUnderControlOf(ev, c) == source.Controller
			}, "Nadier, Agent of the Duskenel — put a +1/+1 counter on it", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
			effects.WhenThisLeaves("Nadier, Agent of the Duskenel — create 1/1 green Elf Warrior tokens equal to its power",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					n := effects.SourcePowerNow(ctx)
					if n <= 0 {
						return nil
					}
					return effects.CreateToken{Template: effects.TokenCard("1/1 green Elf Warrior"), N: n}.Apply(ctx)
				}),
		},
	})
}
