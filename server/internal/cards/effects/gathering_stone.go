package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gathering Stone — Artifact {4} (Lorwyn Eclipsed, #4458):
//
//	"As this artifact enters, choose a creature type.
//	 Spells you cast of the chosen type cost {1} less to cast.
//	 When this artifact enters and at the beginning of your upkeep, look
//	 at the top card of your library. If it's a card of the chosen type,
//	 you may reveal it and put it into your hand. If you don't put the
//	 card into your hand, you may put it into your graveyard."
//
// Herald's Horn widened to every spell and every card of the type: the
// chosen type is ChooseCreatureTypeAsEnters on Card.NamedTribe, and the
// discount is the same battlefield CostModifier without the creature
// clause (a Kindred spell of the type qualifies; the reduction spends
// against generic mana only, per the ruling, which the engine enforces).
//
// The look is one ability with two trigger conditions (Grell
// Philosopher's thisEntersOrYourUpkeep). A matching card is offered to
// the hand first, and revealing is part of the "yes" (the ruling: you
// need not reveal it). Any card that does not go to the hand, a
// non-matching one or one declined, is then offered to the graveyard as
// a plain move (Bison Whistle's putCarriedLibraryCardIntoGraveyard).
// Declining both leaves it on top.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9f6d6e90-7462-46e5-bf2f-725b4badc394",
		Name:         "Gathering Stone",
		Completeness: CompletenessFull,
		AsEnters:     ChooseCreatureTypeAsEnters("Gathering Stone"),
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Spells you cast of the chosen type cost {1} less to cast.",
				YourSpell(), SpellOfTheSourcesChosenType()),
		},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventETB, game.EventBeginUpkeep}, thisEntersOrYourUpkeep,
				"Gathering Stone — look at the top card of your library", gatheringStoneLook),
		},
	})
}

// gatheringStoneLook looks at the top card and offers it to the hand
// when it is a card of the chosen type, or else to the graveyard.
func gatheringStoneLook(g *game.Game, item *game.StackItem) error {
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
	tribe := g.NamedTribeOf(item.SourceCardID)
	if tribe != "" && c.HasSubtype(tribe) {
		return MayChoice{
			Player:   item.Controller,
			Question: "Gathering Stone — reveal " + c.Name + " and put it into your hand?",
			YesLabel: "Put it into your hand",
			NoLabel:  "Don't",
			Then:     gatheringStoneTake,
			Carry:    []uuid.UUID{top},
		}.Apply(ctx)
	}
	return gatheringStoneOfferGraveyard(ctx, top)
}

// gatheringStoneOfferGraveyard is "If you don't put the card into your
// hand, you may put it into your graveyard."
func gatheringStoneOfferGraveyard(ctx *Context, top uuid.UUID) error {
	return MayChoice{
		Player:   ctx.Controller(),
		Question: "Gathering Stone — put the top card of your library into your graveyard?",
		YesLabel: "Put it into your graveyard",
		NoLabel:  "Leave it on top",
		Then:     gatheringStoneToGraveyard,
		Carry:    []uuid.UUID{top},
	}.Apply(ctx)
}

// gatheringStoneTake reveals the card and puts it into the hand on a
// yes; a no goes on to the graveyard offer. Keyed.
var gatheringStoneTake = MayYesNo("confirm/gathering-stone-take",
	func(ctx *Context, r game.ConfirmAnswer) error {
		if !bisonWhistleStillInLibrary(ctx.Game, carryAt(r, 0)) {
			return nil
		}
		ctx.Game.RevealTopOfLibraryForEffect(r.Chooser, ctx.Source(), 1, "Gathering Stone — reveal the top card")
		return ctx.Game.BounceToHandForEffect(carryAt(r, 0))
	},
	func(ctx *Context, r game.ConfirmAnswer) error {
		if !bisonWhistleStillInLibrary(ctx.Game, carryAt(r, 0)) {
			return nil
		}
		return gatheringStoneOfferGraveyard(ctx, carryAt(r, 0))
	})

var gatheringStoneToGraveyard = MayYes("confirm/gathering-stone-to-graveyard", putCarriedLibraryCardIntoGraveyard)
