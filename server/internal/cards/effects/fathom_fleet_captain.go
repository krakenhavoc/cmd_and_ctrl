package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fathom Fleet Captain — Creature — Human Pirate {1}{B}, 2/1:
//
//	"Menace
//	 Whenever this creature attacks, if you control another nontoken
//	 Pirate, you may pay {2}. If you do, create a 2/2 black Pirate
//	 creature token with menace."
//
// The "if" is an intervening one (CR 603.4): checked as the creature
// attacks and again as the trigger resolves, so losing the other Pirate
// in response stops the token. The {2} is asked as it resolves.
func init() {
	Register(Spec{
		OracleID:        "7bcf3090-95c6-4384-a857-9fe571e6da85",
		Name:            "Fathom Fleet Captain",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{
				Name: "you control another nontoken Pirate",
				Test: func(ctx *Context) bool {
					return cmdr2B05ControlsAnotherNontokenPirate(ctx.Game, ctx.Controller(), ctx.Source())
				},
			}, On(game.EventAttack, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return ThisAttacked(ev, source, lki, g) && cmdr2B05ControlsAnotherNontokenPirate(g, source.Controller, source.InstanceID)
			}, "Fathom Fleet Captain — you may pay {2} to create a 2/2 black Pirate creature token with menace",
				func(g *game.Game, item *game.StackItem) error {
					return MayPay{
						Chooser:  item.Controller,
						Cost:     "{2}",
						Question: "Fathom Fleet Captain — pay {2} to create a 2/2 black Pirate creature token with menace?",
						Then:     cmdr2B05FleetCaptainPaid,
					}.Apply(NewContext(g, item))
				})),
		},
	})
}

var cmdr2B05FleetCaptainPaid = PayUnlessBranches("pay-unless/cmdr2-b05-fathom-fleet-captain-pirate",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return CreateToken{Controller: ctx.Controller(), Template: TokenCard("2/2 black Pirate with menace"), N: 1}.Apply(ctx)
	}, nil)
