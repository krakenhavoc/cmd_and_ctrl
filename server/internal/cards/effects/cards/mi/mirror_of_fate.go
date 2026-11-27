package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mirror of Fate — Artifact {5} (Magic 2010, #5585):
//
//	"{T}, Sacrifice this artifact: Choose up to seven face-up exiled cards you
//	 own. Exile all the cards from your library, then put the chosen cards on
//	 top of your library."
//
// The choice is made first, among the face-up cards in exile you own (a card
// exiled face down is not one). The library is then exiled as one event, and
// the chosen cards that are still in exile go on top in an order you choose
// (CR 401.4). The cards exiled from the library are not offered: the choice
// was made before they moved.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "38ffe614-e8f7-4625-b921-9409e7e7ee5b",
		Name:         "Mirror of Fate",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "{T}, Sacrifice this artifact: Choose up to seven face-up exiled cards you own. Exile all the cards from your library, then put the chosen cards on top of your library.",
			Cost:  effects.Plus(effects.TapCost(), effects.SacrificeThis()),
			// ADR 0142: it rearranges your own library; it answers nothing.
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				var offer []uuid.UUID
				if g.Exile != nil {
					for _, c := range g.Exile.Cards {
						if c.Owner == item.Controller && !c.FaceDown {
							offer = append(offer, c.InstanceID)
						}
					}
				}
				g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
					Chooser:  item.Controller,
					Source:   item.SourceCardID,
					Question: "Mirror of Fate — choose up to seven face-up exiled cards you own",
					Cards:    offer,
					Min:      0,
					Max:      7,
					LiveZone: game.ZoneExile,
					ThenKey:  mirrorOfFateChosen,
					Asked:    game.NewAskedItem(item),
				})
				return nil
			},
		}},
	})
}

// mirrorOfFateChosen is the rest of the ability once the cards are chosen:
// exile the library, then stack the chosen cards on top. An on-disk key.
var mirrorOfFateChosen = effects.AskedPick("choose-cards/mirror-of-fate-chosen", func(ctx *effects.Context, r game.CardsChosen) error {
	var library []uuid.UUID
	if p := ctx.Game.PlayerByIDForEffect(r.Chooser); p != nil && p.Library != nil {
		for _, c := range p.Library.Cards {
			library = append(library, c.InstanceID)
		}
	}
	tail := ctx.Tail(mirrorOfFatePutOnTop)
	tail.Carry = append([]uuid.UUID(nil), r.Picked...)
	return ctx.Game.ExileCardsKeyedForEffect(library, tail)
})

// mirrorOfFatePutOnTop puts the chosen cards on top of their owner's library
// in the order the player picks. An on-disk key.
var mirrorOfFatePutOnTop = game.RegisterTail("mirror-of-fate/put-on-top", func(g *game.Game, o game.TailOutcome) error {
	return g.PutInLibraryInChosenOrderThenForEffect(game.PutInLibrarySpec{
		Chooser:   o.Tail.Player,
		Source:    o.Tail.Source,
		Cards:     o.Tail.Carry,
		From:      game.ZoneExile,
		Placement: game.LibraryPlaceTop,
		Reason:    "Mirror of Fate — put the chosen cards on top of your library in any order",
	})
})
