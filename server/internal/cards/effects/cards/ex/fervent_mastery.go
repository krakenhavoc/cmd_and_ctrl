package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fervent Mastery — Sorcery {3}{R}{R} (Strixhaven: School of Mages, #5502):
//
//	"You may pay {2}{R}{R} rather than pay this spell's mana cost.
//	 If the {2}{R}{R} cost was paid, an opponent discards any number of cards,
//	 then draws that many cards.
//	 Search your library for up to three cards, put them into your hand,
//	 shuffle, then discard three cards at random."
//
// Baleful Mastery's bare alternative cost. With it paid, the caster chooses an
// opponent at resolution (ChoosePlayer, as for any "an opponent"); that player
// picks how many cards to discard from the hand they hold and then draws that
// many (Mishra's Command's discard-up-to with its draw). The search waits for
// that to finish (CR 608.2c): it is unconditional, so "up to three" means the
// searcher may take fewer, and the shuffle ends it. The random discard then
// takes three cards at random from the hand as it is after the search (CR
// 701.9b); a hand of fewer than three discards what it has.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cae6d903-3fd6-422d-b21b-92201b78fa1b",
		Name:         "Fervent Mastery",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 3, Discards: 3},
		AlternativeCosts: []game.AlternativeCost{{
			Key:      "fervent-mastery",
			Label:    "Pay {2}{R}{R} rather than pay this spell's mana cost",
			ManaCost: "{2}{R}{R}",
		}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if !ctx.PaidAltCost("fervent-mastery") {
				return ferventMasterySearch(ctx)
			}
			return effects.ChoosePlayer{
				Among:    effects.Opponents,
				Question: "Fervent Mastery — an opponent discards any number of cards, then draws that many cards",
				ThenKey:  ferventMasteryOpponent,
			}.Apply(ctx)
		},
	})
}

// ferventMasteryOpponent is the opponent chosen: they discard any number of
// cards, then draw that many; with no opponent, or an empty hand, the spell
// goes straight on to its search. An on-disk key.
var ferventMasteryOpponent = effects.ChoosePlayerThen("option-pick/fervent-mastery-opponent", func(ctx *effects.Context, opp uuid.UUID, _ []uuid.UUID) error {
	if opp == uuid.Nil {
		return ferventMasterySearch(ctx)
	}
	n := effects.B14HandSize(ctx.Game, opp)
	if n == 0 {
		return ferventMasterySearch(ctx)
	}
	return ctx.Game.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
		Player:   opp,
		Source:   ctx.Source(),
		N:        n,
		UpTo:     true,
		Question: "Fervent Mastery — discard any number of cards, then draw that many",
		Asked:    game.NewAskedItem(ctx.Item),
	}, ferventMasteryDiscarded, nil)
})

// ferventMasteryDiscarded draws the opponent as many cards as they discarded
// and carries on with the caster's search. An on-disk key.
var ferventMasteryDiscarded = game.RegisterDiscardThen("choose-cards/fervent-mastery-discarded", func(g *game.Game, r game.DiscardsDone) error {
	if n := r.Discarded.Count(); n > 0 {
		if err := g.DrawNForEffect(r.Player, n); err != nil {
			return err
		}
	}
	if r.Item == nil {
		return nil
	}
	return ferventMasterySearch(effects.NewContext(g, r.Item))
})

// ferventMasterySearch is "search your library for up to three cards, put
// them into your hand, shuffle, then discard three cards at random".
func ferventMasterySearch(ctx *effects.Context) error {
	return effects.SearchLibrary{
		Player:   ctx.Controller(),
		Dest:     game.ZoneHand,
		Limit:    3,
		UpTo:     true,
		Shuffle:  true,
		Reason:   "Fervent Mastery — up to three cards into your hand",
		ThenTail: ctx.ItemTail(ferventMasteryRandomDiscard),
	}.Apply(ctx)
}

// ferventMasteryRandomDiscard is the last clause, once the search and its
// shuffle are done. An on-disk key.
var ferventMasteryRandomDiscard = effects.CardTail("old-b043/fervent-mastery-random-discard", func(ctx *effects.Context, _ game.TailOutcome) error {
	return effects.DiscardCards{Player: ctx.Controller(), N: 3}.Apply(ctx)
})
