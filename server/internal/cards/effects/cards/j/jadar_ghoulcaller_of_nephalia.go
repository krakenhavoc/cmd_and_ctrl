package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jadar, Ghoulcaller of Nephalia — Legendary Creature — Human Wizard
// {1}{B}, 1/1:
//
//	"At the beginning of your end step, if you control no creatures with
//	 decayed, create a 2/2 black Zombie creature token with decayed."
//
// An intervening "if" (CR 603.4): checked as the end step begins and
// again as the trigger resolves, so a decayed creature that arrives in
// response stops the Zombie. Decayed is read through the layered view, so
// a creature that was granted decayed counts and one that lost it does
// not.
//
// No simplification.
func init() {
	const label = "Jadar, Ghoulcaller of Nephalia — create a 2/2 black Zombie creature token with decayed"
	effects.Register(effects.Spec{
		OracleID:     "fb3defc0-c93a-4423-9ff9-c399f8a8ef1f",
		Name:         "Jadar, Ghoulcaller of Nephalia",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "you control no creatures with decayed",
				Test: func(ctx *effects.Context) bool {
					return !effects.SocB7ControlsADecayedCreature(ctx.Game, ctx.Controller())
				},
			}, effects.On(game.EventBeginEndStep, effects.AllOf(effects.ByYou, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return !effects.SocB7ControlsADecayedCreature(g, source.Controller)
			}), label, effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 black Zombie with decayed"), N: 1}))),
		},
	})
}
