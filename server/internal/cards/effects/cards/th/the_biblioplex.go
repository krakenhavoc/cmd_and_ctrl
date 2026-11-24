package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Biblioplex — Land (Strixhaven: School of Mages, #5502):
//
//	"{T}: Add {C}.
//	 {2}, {T}: Look at the top card of your library. If it's an instant or
//	 sorcery card, you may reveal it and put it into your hand. If you don't
//	 put the card into your hand, you may put it into your graveyard. Activate
//	 only if you have exactly zero or seven cards in hand."
//
// Lantern of Revealing's look-and-decide with the Biblioplex's own rider. The
// hand condition is read when the ability is activated (a CR 602.5b
// restriction on activation, not an intervening "if"). An instant or sorcery on
// top may be revealed and taken (Narset Transcendent's reveal-take); anything
// not taken, whether it was not an instant or sorcery or the player declined,
// may then be put into the graveyard, and otherwise stays on top.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "86ed6073-c35c-4d29-9911-5fe191dd875f",
		Name:         "The Biblioplex",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:     "{2}, {T}: Look at the top card of your library. If it's an instant or sorcery card, you may reveal it and put it into your hand. If you don't put the card into your hand, you may put it into your graveyard. Activate only if you have exactly zero or seven cards in hand.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Condition: oldB51ExactlyZeroOrSevenCardsInHand,
			Effect:    oldB51BiblioplexLook,
		}},
	})
}

func oldB51ExactlyZeroOrSevenCardsInHand(g *game.Game, controller, _ uuid.UUID) bool {
	p := g.PlayerByIDForEffect(controller)
	if p == nil || p.Hand == nil {
		return false
	}
	n := len(p.Hand.Cards)
	return n == 0 || n == 7
}

func oldB51BiblioplexLook(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	top := g.LookAtTopOfLibraryForEffect(item.Controller, 1)
	if len(top) == 0 {
		return nil
	}
	if c, ok := g.LookupCardForEffect(top[0]); ok && (c.IsInstant() || c.IsSorcery()) {
		return effects.TakeFromLibraryToHand{
			Player:   item.Controller,
			Cards:    top,
			Match:    effects.Or(effects.Instant(), effects.Sorcery()),
			Max:      1,
			Optional: true,
			Label:    "The Biblioplex — you may reveal the instant or sorcery card and put it into your hand",
			Key:      oldB51BiblioplexTake,
		}.Apply(ctx)
	}
	return oldB51BiblioplexMayBin(ctx, top[0])
}

// oldB51BiblioplexTake is the take of the instant or sorcery on top: revealed
// when taken, and "if you don't put the card into your hand" asks the
// graveyard question about the card still on top.
var oldB51BiblioplexTake = effects.RegisterTake("take/old-b051-the-biblioplex", effects.TakeDef{
	Reveal: true,
	Then: func(g *game.Game, res effects.TakeFromLibraryResult) error {
		if len(res.Taken) > 0 || len(res.Rest) == 0 {
			return nil
		}
		ctx, _ := effects.AskedContext(g, nil, res.Player, res.Source, nil)
		return oldB51BiblioplexMayBin(ctx, res.Rest[0])
	},
})

func oldB51BiblioplexMayBin(ctx *effects.Context, card uuid.UUID) error {
	return effects.MayChoice{
		Player:   ctx.Controller(),
		Question: "The Biblioplex — put the card into your graveyard?",
		Then:     oldB51BiblioplexBin,
		Carry:    []uuid.UUID{card},
	}.Apply(ctx)
}

var oldB51BiblioplexBin = effects.MayAnswer("confirm/old-b051-the-biblioplex-graveyard", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	if !r.Accepted || len(r.Carry) == 0 {
		return nil
	}
	return effects.RestIntoGraveyard(ctx.Game, r.Carry[:1])
})
