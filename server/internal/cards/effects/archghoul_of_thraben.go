package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Archghoul of Thraben — Creature — Zombie Cleric {2}{B}, 3/2 (Innistrad:
// Crimson Vow):
//
//	"Whenever this creature or another Zombie you control dies, look at the
//	 top card of your library. If it's a Zombie card, you may reveal it and
//	 put it into your hand. If you don't put the card into your hand, you may
//	 put it into your graveyard."
//
// The dies condition is Undead Augur's (the Zombie is read as it last existed
// on the battlefield, CR 603.10a). The look is Gathering Stone's: a Zombie
// card is offered to the hand first, and revealing is part of the "yes"; any
// card that does not go to the hand, a non-Zombie or one declined, is then
// offered to the graveyard as a plain move. Declining both leaves it on top.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "575ffeb4-6d12-4f69-bcc9-15db325daac1",
		Name:         "Archghoul of Thraben",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b17SelfOrZombieYouControlDied(ev, source, g)
			}, "Archghoul of Thraben — look at the top card of your library; if it's a Zombie card, you may reveal it and put it into your hand, otherwise you may put it into your graveyard",
				archghoulLook),
		},
	})
}

func archghoulLook(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	looked := g.LookAtTopOfLibraryForEffect(item.Controller, 1)
	if len(looked) == 0 {
		return nil
	}
	top := looked[0]
	c, ok := g.LookupCardForEffect(top)
	if !ok {
		return nil
	}
	if c.HasSubtype(ctx.TextSubtype("Zombie")) {
		return MayChoice{
			Player:   item.Controller,
			Question: "Archghoul of Thraben — reveal " + c.Name + " and put it into your hand?",
			YesLabel: "Put it into your hand",
			NoLabel:  "Don't",
			Then:     archghoulTake,
			Carry:    []uuid.UUID{top},
		}.Apply(ctx)
	}
	return archghoulOfferGraveyard(ctx, top)
}

// archghoulOfferGraveyard is "If you don't put the card into your hand, you
// may put it into your graveyard."
func archghoulOfferGraveyard(ctx *Context, top uuid.UUID) error {
	return MayChoice{
		Player:   ctx.Controller(),
		Question: "Archghoul of Thraben — put the top card of your library into your graveyard?",
		YesLabel: "Put it into your graveyard",
		NoLabel:  "Leave it on top",
		Then:     archghoulToGraveyard,
		Carry:    []uuid.UUID{top},
	}.Apply(ctx)
}

// archghoulTake reveals the card and puts it into the hand on a yes; a no goes
// on to the graveyard offer. An on-disk key.
var archghoulTake = MayYesNo("confirm/archghoul-of-thraben-take",
	func(ctx *Context, r game.ConfirmAnswer) error {
		if !bisonWhistleStillInLibrary(ctx.Game, carryAt(r, 0)) {
			return nil
		}
		ctx.Game.RevealTopOfLibraryForEffect(r.Chooser, ctx.Source(), 1, "Archghoul of Thraben — reveal the top card")
		return ctx.Game.BounceToHandForEffect(carryAt(r, 0))
	},
	func(ctx *Context, r game.ConfirmAnswer) error {
		if !bisonWhistleStillInLibrary(ctx.Game, carryAt(r, 0)) {
			return nil
		}
		return archghoulOfferGraveyard(ctx, carryAt(r, 0))
	})

// archghoulToGraveyard is the graveyard offer's yes. An on-disk key.
var archghoulToGraveyard = MayYes("confirm/archghoul-of-thraben-to-graveyard", putCarriedLibraryCardIntoGraveyard)
