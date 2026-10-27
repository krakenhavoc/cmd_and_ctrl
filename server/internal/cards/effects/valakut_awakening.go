package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Valakut Awakening // Valakut Stoneforge — modal double-faced card.
// This file is the FRONT face, Instant {2}{R}:
//
//	"Put any number of cards from your hand on the bottom of your
//	 library, then draw that many cards plus one."
//
// The back face, Valakut Stoneforge, is a plain "enters tapped, taps
// for {R}" land and is already registered with the rest of the MDFC
// land cycle in mdfc_lands.go.
//
// Fable of the Mirror-Breaker's chapter II ("discard up to two, then
// draw that many") is the same SHAPE — a from-hand pick, then a draw
// sized off how many were picked — with two differences the card's
// text demands: the pile goes to the BOTTOM OF THE LIBRARY rather than
// the graveyard (PutInLibraryInChosenOrderThenForEffect,
// LibraryPlaceBottom, library_order.go), and the count is uncapped
// ("any number", floor zero, ceiling the whole hand) with the draw
// count one higher than what was put back.
//
// The order the put-back cards land in is a real decision (CR 701.19a:
// the player doing the placing chooses, absent a printed "in any
// order" clause saying otherwise) and gets its own prompt for two or
// more cards; a single card needs no ordering and a decline needs no
// placement prompt at all.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ff0ab867-b710-4b1a-baed-95fc3cf68f79",
		Name:         "Valakut Awakening",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return valakutAwakeningPutThenDraw(ctx, item.Controller)
		},
	})
}

// valakutAwakeningPutThenDraw is "Put any number of cards from your
// hand on the bottom of your library, then draw that many cards plus
// one."
func valakutAwakeningPutThenDraw(ctx *Context, player uuid.UUID) error {
	hand := allHandCardIDs(ctx.Game, player)
	if len(hand) == 0 {
		return ctx.Game.DrawNForEffect(player, 1)
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   ctx.Source(),
		Question: "Valakut Awakening — put any number of cards from your hand on the bottom of your library",
		Cards:    hand,
		Min:      0,
		Max:      len(hand),
		ThenKey:  valakutAwakeningPicked,
		LiveZone: game.ZoneHand,
	})
	return nil
}

// valakutAwakeningPicked is the any-number put-back, keyed (#4253): the
// player is the chooser and the card the source, so the prompt carries
// nothing else. The order prompt it opens for two or more cards is
// PutInLibrarySpec's, whose "then" is the keyed draw (thenDrawCards).
var valakutAwakeningPicked = game.RegisterChooseCardsThen("choose-cards/valakut-awakening-put-back",
	func(g *game.Game, r game.CardsChosen) error {
		player, picked := r.Chooser, r.Picked
		if len(picked) == 0 {
			return g.DrawNForEffect(player, 1)
		}
		return g.PutInLibraryInChosenOrderThenForEffect(game.PutInLibrarySpec{
			Chooser:   player,
			Source:    r.Source,
			Cards:     picked,
			From:      game.ZoneHand,
			Placement: game.LibraryPlaceBottom,
			Reason:    "Valakut Awakening — order the cards on the bottom of your library",
			ThenTail:  game.KeyedTail{Ref: thenDrawCards, Player: player, N: len(picked) + 1},
		})
	})
