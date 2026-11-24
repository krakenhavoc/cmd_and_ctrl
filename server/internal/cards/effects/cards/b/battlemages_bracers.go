package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Battlemage's Bracers — Artifact — Equipment {2}{R} (Commander 2021, S82 OLD
// b058, #5504):
//
//	"Equipped creature has haste.
//	 Whenever an ability of equipped creature is activated, if it isn't a
//	 mana ability, you may pay {1}. If you do, copy that ability. You may
//	 choose new targets for the copy.
//	 Equip {2}"
//
// Illusionist's Bracers' trigger (an ability of the equipped creature is
// activated; only a CR 602 activation emits EventActivateAbility, so a mana
// ability never reaches it) with Rings of Brighthearth's optional payment: the
// {1} is a resolution-time MayPay, and on payment the activation named by the
// triggering event, still on the stack beneath the trigger, is copied under
// the Bracers' controller with the new-targets offer (CR 707.10c). The
// equipped creature has haste through the attached-grant static.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bd56225e-8c80-4328-9ca0-0fff4969573d",
		Name:         "Battlemage's Bracers",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.GrantToAttached("haste")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventActivateAbility, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source.IsAttachedTo(ev.CardID)
			}, "Battlemage's Bracers — you may pay {1} to copy that ability", battlemagesBracersOffer),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipAbility("{2}"),
		},
	})
}

// battlemagesBracersOffer asks for the {1} as the trigger resolves.
func battlemagesBracersOffer(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if ctx.Trigger().Event.StackItemID == uuid.Nil {
		return nil
	}
	return effects.MayPay{
		Chooser:  item.Controller,
		Cost:     "{1}",
		Question: "Battlemage's Bracers — pay {1} to copy that ability?",
		Then:     battlemagesBracersPaid,
	}.Apply(ctx)
}

// battlemagesBracersPaid is the paid half, keyed: the ability is the asking
// trigger's event. An on-disk key.
var battlemagesBracersPaid = effects.PayUnlessBranches("pay-unless/battlemages-bracers-copy-ability",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.CopyAbility{
			ItemID:           ctx.Trigger().Event.StackItemID,
			Controller:       ctx.Controller(),
			ChooseNewTargets: true,
		}.Apply(ctx)
	}, nil)
