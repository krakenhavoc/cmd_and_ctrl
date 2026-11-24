package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Amareth, the Lustrous — Legendary Creature — Dragon {3}{G}{W}{U}, 6/6:
//
//	"Flying
//	 Whenever another permanent you control enters, look at the top card of
//	 your library. If it shares a card type with that permanent, you may reveal
//	 that card and put it into your hand."
//
// Herald's Horn's look-then-offer, keyed to the permanent that entered. The
// entering permanent is read at resolution as it last existed (CR 608.2h,
// Context.TriggeringPermanent), so one that has already left is still
// compared; the comparison is on card types only (CR 205.2a), never subtypes
// or supertypes. The look is private; the reveal happens only on a yes. A top
// card that shares nothing is looked at and left where it is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8cccc8d0-bdd8-4853-8330-6f6a57695d11",
		Name:            "Amareth, the Lustrous",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := effects.EnteredUnderYourControl(ev, source, g, true)
				return ok
			}, "Amareth, the Lustrous — look at the top card of your library; if it shares a card type with that permanent, you may reveal it and put it into your hand",
				amarethLook),
		},
	})
}

func amarethLook(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	entered, known := ctx.TriggeringPermanent()
	looked := g.LookAtTopOfLibraryForEffect(item.Controller, 1)
	if !known || len(looked) == 0 {
		return nil
	}
	top := looked[0]
	c, ok := g.LookupCardForEffect(top)
	if !ok || !effects.SharesACardTypeWith(entered, c) {
		return nil
	}
	return effects.MayChoice{
		Player:   item.Controller,
		Question: "Amareth, the Lustrous — reveal " + c.Name + " and put it into your hand?",
		Then:     amarethTake,
		Carry:    []uuid.UUID{top},
	}.Apply(ctx)
}

// amarethTake is the "yes": reveal the top card and put it into the hand.
// Keyed (#4291).
var amarethTake = effects.MayYes("confirm/amareth-the-lustrous-take", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	ctx.Game.RevealTopOfLibraryForEffect(r.Chooser, ctx.Source(), 1, "Amareth, the Lustrous — reveal the top card")
	return ctx.Game.BounceToHandForEffect(effects.CarryAt(r, 0))
})
