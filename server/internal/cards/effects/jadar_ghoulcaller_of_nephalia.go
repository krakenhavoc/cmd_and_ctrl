package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "fb3defc0-c93a-4423-9ff9-c399f8a8ef1f",
		Name:         "Jadar, Ghoulcaller of Nephalia",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{
				Name: "you control no creatures with decayed",
				Test: func(ctx *Context) bool { return !socB7ControlsADecayedCreature(ctx.Game, ctx.Controller()) },
			}, On(game.EventBeginEndStep, AllOf(ByYou, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return !socB7ControlsADecayedCreature(g, source.Controller)
			}), label, Do(CreateToken{Template: TokenCard("2/2 black Zombie with decayed"), N: 1}))),
		},
	})
}
