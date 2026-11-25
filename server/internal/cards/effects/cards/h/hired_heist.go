package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hired Heist — Conspiracy (Conspiracy: Take the Crown, S82 OLD batch 269,
// #5540):
//
//	"Hidden agenda
//	 Whenever a creature you control with the chosen name deals combat damage
//	 to a player, you may pay {U}. If you do, draw a card."
//
// Hidden agenda (ADR 0227): face down it does nothing (CR 315.5). One trigger
// per creature that dealt the damage (it is not a "one or more" trigger), the
// name read from the damage source as it last was on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b0d84455-dc35-45d3-901e-8d28363142f4",
		Name:         "Hired Heist",
		Completeness: effects.CompletenessFull,
		HiddenAgenda: effects.HiddenAgendaOneName,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, hiredHeistDamage,
				"Hired Heist — you may pay {U}; if you do, draw a card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{U}",
						Question: "Hired Heist — pay {U} to draw a card?",
						Then:     hiredHeistPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

func hiredHeistDamage(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !effects.CombatDamageToPlayerBy(ev, source.Controller, g) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.Source)
	return ok && effects.HasTheChosenName(c, source)
}

var hiredHeistPaid = effects.PayUnlessBranches("pay-unless/hired-heist-draw",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error { return effects.DrawCards{N: 1}.Apply(ctx) }, nil)
