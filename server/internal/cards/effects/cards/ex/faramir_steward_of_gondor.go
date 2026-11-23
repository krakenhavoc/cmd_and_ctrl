package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Faramir, Steward of Gondor — Legendary Creature — Human Noble {1}{W}{U},
// 2/2 (Tales of Middle-earth Commander):
//
//	"Whenever a legendary creature you control with mana value 4 or
//	 greater enters, you become the monarch.
//	 At the beginning of your end step, if you're the monarch, create two
//	 1/1 white Human Soldier creature tokens."
//
// The entry clause reads the entering creature through the layered view
// and its mana value as a permanent (a token or copy of a legend counts by
// the mana value it copies, CR 202.3). The end-step clause has an
// intervening "if" (CR 603.4): it is checked as the step begins and again
// as the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "466af6da-52e8-4ac5-b659-7492e3594ff8",
		Name:         "Faramir, Steward of Gondor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
				if !ok || !c.IsCreature() || !c.IsLegendary() {
					return false
				}
				mv, ok := g.ManaValueForEffect(c)
				return ok && mv >= 4
			}, "Faramir, Steward of Gondor — you become the monarch", effects.Do(effects.BecomeTheMonarch{})),
			effects.InterveningIf(effects.CondMonarch{},
				effects.AtYourEndStep("Faramir, Steward of Gondor — create two 1/1 white Human Soldier creature tokens",
					effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Human Soldier"), N: 2}))),
		},
	})
}
