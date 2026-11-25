package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grakmaw, Skyclave Ravager — Legendary Creature — Hydra Horror {1}{B}{G},
// 0/0 (Zendikar Rising, #5508):
//
//	"Grakmaw enters with three +1/+1 counters on it.
//	 Whenever another creature you control dies, if it had a +1/+1 counter
//	 on it, put a +1/+1 counter on Grakmaw.
//	 When Grakmaw dies, create an X/X black and green Hydra creature token,
//	 where X is the number of +1/+1 counters on Grakmaw."
//
// The entry counters are a self-replacement (CR 614.1c). The second ability
// has an intervening "if" (CR 603.4): the creature must have had a +1/+1
// counter as it last existed (CR 603.10a) and be controlled by Grakmaw's
// controller as it died. X is the +1/+1 counters Grakmaw had as it died
// (CR 608.2h); the Hydra's size is set as it is created, so it is a 0/0
// template resized to X/X.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9c93acf3-d3da-4b4b-947f-ec8fc99d0d34",
		Name:         "Grakmaw, Skyclave Ravager",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(game.CounterPlusOne, 3, "Grakmaw enters with three +1/+1 counters on it"),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.S82OldB097AnotherCreatureWithAPlusOneCounterYouControlDied(ev, source, g)
			}, "Grakmaw, Skyclave Ravager — put a +1/+1 counter on Grakmaw",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
			effects.WhenThisDies("Grakmaw, Skyclave Ravager — create an X/X black and green Hydra, X being the +1/+1 counters on Grakmaw",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					info, ok := ctx.SourcePermanent()
					if !ok {
						return nil
					}
					x := info.Counters[game.CounterPlusOne]
					tok := ctx.TextToken(effects.TokenCard("0/0 black and green Hydra"))
					tok.Power, tok.Toughness = x, x
					return ctx.Game.CreateTokenForEffect(item.Controller, tok, 1)
				}),
		},
	})
}
