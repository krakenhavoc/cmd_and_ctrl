package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thran Tome — Artifact — Book {4} (Weatherlight, S82 older sets batch 777, #5629):
//
//	"{5}, {T}: Reveal the top three cards of your library. Target opponent chooses one of those
//	 cards. Put that card into your graveyard, then draw two cards."
//
// Murmurs from Beyond's reveal and choice, with the unchosen cards drawn: the target opponent
// (picked on activation) chooses one of the revealed cards, the controller looking on; it goes to the
// graveyard, then two cards are drawn from the top (the other two, when three were revealed).
// A library of fewer than three reveals what it has. A target that has become illegal does nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "78d2a3fd-864c-49f3-b531-c71dc3b35977",
		Name:         "Thran Tome",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}, {T}: Reveal the top three cards of your library. Target opponent chooses one of those cards. Put that card into your graveyard, then draw two cards.",
			Cost:    effects.Plus(effects.ManaCost("{5}"), effects.TapCost()),
			Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
			Purpose: game.Purpose{Draws: 2},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				opp := effects.FirstLegalTargetPlayer(ctx)
				if opp == uuid.Nil {
					return nil
				}
				ids := g.RevealTopOfLibraryForEffect(item.Controller, item.SourceCardID, 3, "Thran Tome")
				if len(ids) == 0 {
					return nil
				}
				g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
					Chooser:    opp,
					FromPlayer: ctx.Controller(),
					Source:     ctx.Source(),
					Question:   "Thran Tome — choose one of the revealed cards: it goes to their graveyard, then they draw two cards",
					Cards:      ids,
					Min:        1,
					Max:        1,
					LiveZone:   game.ZoneLibrary,
					ThenKey:    thranTomeChosen,
					Asked:      game.NewAskedItem(ctx.Item),
				})
				return nil
			},
		}},
	})
}

// thranTomeChosen puts the chosen card into the graveyard, then draws two. An on-disk key.
var thranTomeChosen = effects.AskedPick("choose-cards/s82-old-b777-thran-tome-chosen", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	if err := effects.RestIntoGraveyard(ctx.Game, r.Picked[:1]); err != nil {
		return err
	}
	return ctx.Game.DrawNForEffect(ctx.Controller(), 2)
})
