package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sorcerer's Broom — Artifact Creature — Spirit {2}, 2/1 (Throne of Eldraine,
// S82 old sets):
//
//	"Whenever you sacrifice another permanent, you may pay {3}. If you do,
//	 create a token that's a copy of this creature."
//
// Copy Catchers' pay-then-copy: the sacrifice trigger (tokens included, the
// Broom itself excluded) asks for {3} as it resolves, and the token is a copy
// of the Broom's copiable values, or its last-known ones if it has left
// (CR 707.2, 608.2h). Each copy has the trigger too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e73efb04-30ac-4484-b11b-e98f0353b9b6",
		Name:         "Sorcerer's Broom",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventSacrifice, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventSacrifice && ev.Actor == source.Controller && ev.CardID != source.InstanceID
			}, "Sorcerer's Broom — you may pay {3} to create a token that's a copy of this creature",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{3}",
						Question: "Sorcerer's Broom — pay {3} to create a token that's a copy of this creature?",
						Then:     sorcerersBroomPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// sorcerersBroomPaid is "create a token that's a copy of this creature", once
// {3} is paid. An on-disk key: never renamed, never reused.
var sorcerersBroomPaid = effects.PayUnlessBranches("pay-unless/sorcerers-broom-copy",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.CreateTokenCopy{Controller: ctx.Controller(), Copy: ctx.Source(), N: 1}.Apply(ctx)
	}, nil)
