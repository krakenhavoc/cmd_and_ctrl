package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bess, Soul Nourisher — Legendary Creature — Human Citizen {1}{G}{W}, 1/1
// (Streets of New Capenna Commander):
//
//	"Whenever one or more other creatures you control with base power and
//	 toughness 1/1 enter, put a +1/+1 counter on Bess.
//	 Whenever Bess attacks, each other creature you control with base power
//	 and toughness 1/1 gets +X/+X until end of turn, where X is the number
//	 of +1/+1 counters on Bess."
//
// Base power and toughness is the value after layer 7b (CR 208.4b), read as
// the creature enters. Creatures entering together are one trigger
// (OncePerBatch), so Bess grows once however many arrive. The attack pump
// fixes who is affected as the trigger resolves (CR 611.2c) and reads X
// then.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "df72b5d6-6f9e-4d6b-acf6-7dec4ff35468",
		Name:         "Bess, Soul Nourisher",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, true)
				return ok && effects.BaseOneByOne(g, source.Controller, c)
			}, "Bess, Soul Nourisher — put a +1/+1 counter on Bess",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}))),
			effects.WheneverThisAttacks("Bess, Soul Nourisher — each other creature you control with base power and toughness 1/1 gets +X/+X until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					src, ok := g.LookupCardForEffect(ctx.Source())
					if !ok {
						return nil
					}
					x := src.Counters[game.CounterPlusOne]
					if x <= 0 {
						return nil
					}
					return effects.BoostUntilEOT{
						Match:     effects.And(effects.YouControl(), effects.NotSelf(src.InstanceID), effects.BaseOneByOne),
						Power:     x,
						Toughness: x,
						Label:     "Bess, Soul Nourisher — +X/+X until end of turn",
					}.Apply(ctx)
				}),
		},
	})
}
