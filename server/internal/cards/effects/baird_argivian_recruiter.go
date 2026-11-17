package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Baird, Argivian Recruiter — Legendary Creature — Human Soldier {R}{W}, 2/2
// (Dominaria United, S82 DMU batch 4):
//
//	"At the beginning of your end step, if you control a creature with
//	 power greater than its base power, create a 1/1 white Soldier creature
//	 token."
//
// Kutzil's base-power comparison (CR 208.4b) as an intervening if (CR 603.4),
// checked as the end step begins and again as the trigger resolves, over
// every creature the controller controls (Baird included).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1d078108-f88c-4a32-a28f-bae5182b2f53",
		Name:         "Baird, Argivian Recruiter",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{
				Name: "you control a creature with power greater than its base power",
				Test: func(ctx *Context) bool { return dmuB04ControlsCreatureWithPowerOverBase(ctx.Game, ctx.Controller()) },
			}, On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && dmuB04ControlsCreatureWithPowerOverBase(g, source.Controller)
			}, "Baird — create a 1/1 white Soldier creature token",
				Do(CreateToken{Template: TokenCard("1/1 white Soldier"), N: 1}))),
		},
	})
}
