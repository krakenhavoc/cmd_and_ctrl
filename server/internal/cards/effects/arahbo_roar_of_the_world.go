package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Arahbo, Roar of the World — Legendary Creature — Cat Avatar
// {3}{G}{W}, 5/5 (EDHREC rank 7833):
//
//	"Eminence — At the beginning of combat on your turn, if Arahbo is in
//	 the command zone or on the battlefield, another target Cat you
//	 control gets +3/+3 until end of turn.
//	 Whenever another Cat you control attacks, you may pay {1}{G}{W}. If
//	 you do, it gains trample and gets +X/+X until end of turn, where X
//	 is its power."
//
// The eminence line works from the command zone (#2802,
// EminenceTrigger); "another" is Arahbo itself, so from the command
// zone every Cat its owner controls is a legal target. With no other
// Cat the trigger is removed as it would go on the stack (CR 603.3d).
//
// The attack trigger is per Cat ("another Cat you control attacks", not
// "one or more"), and the payment is made as it resolves. X is read
// then, once, from the attacker's power with every effect and counter
// on it, so a later pump is not counted again; a Cat that is no longer
// on the battlefield gets nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "66944a11-40a1-4f3a-9f83-52324e0edfef",
		Name:         "Arahbo, Roar of the World",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			EminenceTrigger(Targeting(AtBeginningOfYourCombat(
				"Arahbo, Roar of the World — another target Cat you control gets +3/+3",
				func(g *game.Game, item *game.StackItem) error {
					return vsPumpTheTarget(NewContext(g, item), 3, 3, "Arahbo, Roar of the World — +3/+3")
				}),
				Another(TargetCreature("another target Cat you control", OfCreatureType("Cat"), YouControl())))),
			On(game.EventAttack, anotherCatYouControlAttacked,
				"Arahbo, Roar of the World — you may pay {1}{G}{W} to give that Cat trample and +X/+X",
				arahboMayPayToPump),
		},
	})
}

// anotherCatYouControlAttacked is "Whenever another Cat you control
// attacks".
func anotherCatYouControlAttacked(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return ev.CardID != source.InstanceID && attackDeclaredByYou(ev, source.Controller) && attackerHasSubtype(g, ev, "Cat")
}

// arahboMayPayToPump is the attack trigger's body: the offer, then
// trample and +X/+X on the attacker if it was paid.
func arahboMayPayToPump(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	cat := item.Trigger.Event.CardID
	return MayPay{
		Chooser:  item.Controller,
		Cost:     "{1}{G}{W}",
		Question: "Arahbo, Roar of the World — pay {1}{G}{W}: the attacking Cat gains trample and gets +X/+X, where X is its power?",
		OnPay: func(ctx *Context) error {
			return arahboPump(ctx, cat)
		},
	}.Apply(NewContext(g, item))
}

// arahboPump is "it gains trample and gets +X/+X until end of turn,
// where X is its power": one effect at one timestamp, X read now.
func arahboPump(ctx *Context, cat uuid.UUID) error {
	x, _, ok := doubledBy(ctx.Game, cat, false)
	if !ok {
		return nil
	}
	return untilEndOfTurn(ctx, cat, nil, "Arahbo, Roar of the World — trample and +X/+X",
		game.ModifyPTMod(x, x), game.AddKeywordsMod("trample"))
}
