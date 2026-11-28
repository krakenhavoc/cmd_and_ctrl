package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Credit Voucher — Artifact {2} (Mercadian Masques, S82 OLD, #5619):
//
//	"{2}, {T}, Sacrifice this artifact: Shuffle any number of cards from your
//	 hand into your library, then draw that many cards."
//
// The sacrifice is a cost, so the Voucher is gone whatever is chosen. The
// controller picks any number of cards from their hand, none included, as the
// ability resolves; those cards are put into the library together, the library
// is shuffled and as many cards are drawn as went in (Winds of Change's
// shuffle and redraw, over a chosen set).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8496a8f1-f505-403a-8007-c35aedb90e2d",
		Name:         "Credit Voucher",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice this artifact: Shuffle any number of cards from your hand into your library, then draw that many cards.",
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.SacrificeThis()),
			Effect:  creditVoucherChoose,
		}},
	})
}

// creditVoucherChoose asks the controller for any number of cards from
// their hand (none included) to shuffle into the library.
func creditVoucherChoose(g *game.Game, item *game.StackItem) error {
	hand := effects.AllHandCardIDs(g, item.Controller)
	if len(hand) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Credit Voucher — shuffle any number of cards from your hand into your library, then draw that many cards",
		Cards:    hand,
		Min:      0,
		Max:      len(hand),
		LiveZone: game.ZoneHand,
		ThenKey:  creditVoucherCreditVoucherPicked,
	})
	return nil
}

// creditVoucherCreditVoucherPicked tucks the chosen cards together; the shuffle and
// the draw follow in Winds of Change's tail (windsOfChangeShuffled), which
// counts what reached the library. An on-disk key: never renamed, never reused.
var creditVoucherCreditVoucherPicked = game.RegisterChooseCardsThen("choose-cards/credit-voucher-shuffle", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return g.TuckCardsToLibraryKeyedForEffect(r.Picked, game.TuckOptions{},
		game.KeyedTail{Ref: effects.WindsOfChangeShuffled, Player: r.Chooser, Carry: []uuid.UUID{r.Chooser}})
})
