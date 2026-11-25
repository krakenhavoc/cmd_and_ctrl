package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Adriana's Valor — Conspiracy (Conspiracy: Take the Crown, S82 OLD batch
// 267, #5540):
//
//	"Hidden agenda
//	 Whenever a creature you control with the chosen name attacks, you may
//	 pay {W}. If you do, that creature gains indestructible until end of turn."
//
// Hidden agenda (ADR 0227). The attacker is the triggering event's card;
// the mana is paid only if the player can pay it (a payment they can't make
// is a decline), and the creature gains indestructible through the shared
// until-end-of-turn keyword grant, pinned to that creature (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "512a9ab9-ddd4-4872-b4f4-dee56f9f5575",
		Name:         "Adriana's Valor",
		Completeness: effects.CompletenessFull,
		HiddenAgenda: effects.HiddenAgendaOneName,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, adrianasValorAttacker,
				"Adriana's Valor — you may pay {W}; if you do, that creature gains indestructible until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{W}",
						Question: "Adriana's Valor — pay {W} to give that creature indestructible until end of turn?",
						Then:     adrianasValorPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

func adrianasValorAttacker(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventAttack {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && c.Controller == source.Controller && effects.HasTheChosenName(c, source)
}

var adrianasValorPaid = effects.PayUnlessBranches("pay-unless/adrianas-valor-indestructible",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.GrantKeywordUntilEOT{Target: ctx.Item.Trigger.Event.CardID, Keywords: []string{"indestructible"}}.Apply(ctx)
	}, nil)
