package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Traveling Botanist — Creature — Dog Scout {1}{G}, 2/3 (Tarkir:
// Dragonstorm, #4463):
//
//	"Whenever this creature becomes tapped, look at the top card of your
//	 library. If it's a land card, you may reveal it and put it into your
//	 hand. If you don't put the card into your hand, you may put it into
//	 your graveyard."
//
// "Becomes tapped" is ThisBecameTapped over both tap events (an attacker
// is tapped with no EventTapCard; a vigilance attacker did not become
// tapped). The look is Bison Whistle's, private to you. A land card is
// offered to your hand first (revealed to the table as it moves); the
// second sentence is the else of that offer, so it covers a non-land card
// and a land you declined alike, and is a plain move to the graveyard
// (not a mill). Declining both leaves the card on top, and a card that
// left the library before an answer is a new object (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "14934ab0-41ab-44f8-bb28-2b0457eb6f5f",
		Name:         "Traveling Botanist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, ThisBecameTapped,
				"Traveling Botanist — look at the top card of your library",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					looked := g.LookAtTopOfLibraryForEffect(item.Controller, 1)
					if len(looked) == 0 {
						return nil
					}
					c, ok := g.LookupCardForEffect(looked[0])
					if !ok {
						return nil
					}
					if !c.IsLand() {
						return travelingBotanistOfferGraveyard(ctx, looked[0])
					}
					return MayChoice{
						Question: "Traveling Botanist — reveal " + c.Name + " and put it into your hand?",
						YesLabel: "Reveal it and put it into your hand",
						NoLabel:  "Don't put it into your hand",
						Then:     travelingBotanistTakeLand,
						Carry:    []uuid.UUID{looked[0]},
					}.Apply(ctx)
				}),
		},
	})
}

// travelingBotanistOfferGraveyard is the second sentence: you may put the
// looked-at card into your graveyard.
func travelingBotanistOfferGraveyard(ctx *Context, top uuid.UUID) error {
	return MayChoice{
		Question: "Traveling Botanist — put the top card of your library into your graveyard?",
		YesLabel: "Put it into your graveyard",
		NoLabel:  "Leave it on top",
		Then:     bisonWhistleToGraveyard,
		Carry:    []uuid.UUID{top},
	}.Apply(ctx)
}

// travelingBotanistTakeLand is the answer to the hand offer: yes reveals
// the card and takes it (if it is still on top of the library); no, or a
// card that is no longer there, goes on to the graveyard offer when the
// card is still in the library.
var travelingBotanistTakeLand = MayYesNo("std-b9/traveling-botanist-take-land",
	func(ctx *Context, r game.ConfirmAnswer) error {
		if !bisonWhistleStillInLibrary(ctx.Game, carryAt(r, 0)) {
			return nil
		}
		ctx.Game.RevealTopOfLibraryForEffect(r.Chooser, ctx.Source(), 1, "Traveling Botanist — reveal the top card")
		return ctx.Game.BounceToHandForEffect(carryAt(r, 0))
	},
	func(ctx *Context, r game.ConfirmAnswer) error {
		if !bisonWhistleStillInLibrary(ctx.Game, carryAt(r, 0)) {
			return nil
		}
		return travelingBotanistOfferGraveyard(ctx, carryAt(r, 0))
	})
