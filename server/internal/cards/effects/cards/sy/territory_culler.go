package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Territory Culler — Creature — Eldrazi {4}{G}, 7/5 (Modern Horizons 3):
//
//	"Devoid
//	 Reach
//	 Landfall — Whenever a land you control enters, look at the top card of
//	 your library. If it's a creature card, you may reveal it and put it into
//	 your hand. If you don't put the card into your hand, you may put it into
//	 your graveyard."
//
// Gathering Stone's shape: the top card is looked at as the ability resolves;
// only a creature card offers the hand, and a card left there is offered for the
// graveyard whatever it is.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a075a37b-beb6-409d-b5b6-d291921cdef0",
		Name:            "Territory Culler",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid, "reach"},
		Triggered: []game.TriggeredAbility{
			effects.Landfall("Territory Culler — look at the top card of your library", gapB51CullerLook),
		},
	})
}

func gapB51CullerLook(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	looked := g.LookAtTopOfLibraryForEffect(item.Controller, 1)
	if len(looked) == 0 {
		return nil
	}
	top := looked[0]
	c, ok := g.LookupCardForEffect(top)
	if !ok {
		return nil
	}
	if c.IsCreature() {
		return effects.MayChoice{
			Player:   item.Controller,
			Question: "Territory Culler — reveal " + c.Name + " and put it into your hand?",
			YesLabel: "Put it into your hand",
			NoLabel:  "Don't",
			Then:     gapB51CullerTake,
			Carry:    []uuid.UUID{top},
		}.Apply(ctx)
	}
	return gapB51CullerOfferGraveyard(ctx, top)
}

func gapB51CullerOfferGraveyard(ctx *effects.Context, top uuid.UUID) error {
	return effects.MayChoice{
		Player:   ctx.Controller(),
		Question: "Territory Culler — put the top card of your library into your graveyard?",
		YesLabel: "Put it into your graveyard",
		NoLabel:  "Leave it on top",
		Then:     gapB51CullerToGraveyard,
		Carry:    []uuid.UUID{top},
	}.Apply(ctx)
}

var gapB51CullerTake = effects.MayYesNo("confirm/territory-culler-take",
	func(ctx *effects.Context, r game.ConfirmAnswer) error {
		if !effects.BisonWhistleStillInLibrary(ctx.Game, effects.CarryAt(r, 0)) {
			return nil
		}
		ctx.Game.RevealTopOfLibraryForEffect(r.Chooser, ctx.Source(), 1, "Territory Culler — reveal the top card")
		return ctx.Game.BounceToHandForEffect(effects.CarryAt(r, 0))
	},
	func(ctx *effects.Context, r game.ConfirmAnswer) error {
		if !effects.BisonWhistleStillInLibrary(ctx.Game, effects.CarryAt(r, 0)) {
			return nil
		}
		return gapB51CullerOfferGraveyard(ctx, effects.CarryAt(r, 0))
	})

var gapB51CullerToGraveyard = effects.MayYes("confirm/territory-culler-to-graveyard", effects.PutCarriedLibraryCardIntoGraveyard)
