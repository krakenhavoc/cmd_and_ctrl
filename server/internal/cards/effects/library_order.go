package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// library_order.go — the card-side sentences over ADR 0088's ordered
// library placement (#996).
//
// The engine owns the prompt, the move and who knows what afterwards
// (game/library_order.go). What a card prints is which cards, which
// lane, and who chooses; that is all these carry.

// PutInLibraryInAnyOrder is "put them on top of / on the bottom of your
// library in any order", and "put it on your choice of the top or
// bottom of its owner's library".
//
// Like Scry it only QUEUES: anything the card does after the placement
// goes in Then, not on the next line, or it runs before the player has
// answered.
type PutInLibraryInAnyOrder struct {
	// Chooser orders the pile. Zero means the resolving item's
	// controller.
	Chooser uuid.UUID

	// Cards are the cards to place, top-first when they are already in
	// a library. Each goes to its OWNER's library.
	Cards []uuid.UUID

	// From is the zone kind the cards are in now. A card that has left
	// it is skipped (CR 400.7). Empty accepts any zone.
	From game.ZoneKind

	// Placement is the lane the card prints: game.LibraryPlaceTop,
	// game.LibraryPlaceBottom or game.LibraryPlaceTopOrBottom.
	Placement game.LibraryPlacement

	// Label is the prompt banner, "<card> — <the printed clause>".
	Label string

	// Then is the rest of the effect. Runs once the pile is placed.
	Then func(g *game.Game) error
}

func (p PutInLibraryInAnyOrder) Apply(ctx *Context) error {
	chooser := p.Chooser
	if chooser == uuid.Nil {
		chooser = ctx.Controller()
	}
	return ctx.Game.PutInLibraryInChosenOrderThenForEffect(game.PutInLibrarySpec{
		Chooser:   chooser,
		Source:    ctx.Source(),
		Cards:     p.Cards,
		From:      p.From,
		Placement: p.Placement,
		Reason:    p.Label,
		Then:      p.Then,
	})
}

// TakeRestOnBottomInAnyOrder is the Then for "put the rest on the
// bottom of your library IN ANY ORDER" — Impulse, Goblin Ringleader,
// Augur of Bolas and the forty-odd cards that print it (ADR 0088).
// TakeRestOnBottomInRandomOrder's twin, for the cards that print the
// player's choice rather than a random order.
func TakeRestOnBottomInAnyOrder(g *game.Game, res TakeFromLibraryResult) error {
	return g.PutInLibraryInChosenOrderThenForEffect(game.PutInLibrarySpec{
		Chooser:   res.Player,
		Source:    res.Source,
		Cards:     res.Rest,
		From:      game.ZoneLibrary,
		Placement: game.LibraryPlaceBottom,
		Reason:    libraryOrderLabel(g, res.Source, "put the rest on the bottom of your library in any order"),
	})
}

// LookAtTopThenTakeOneRestOnBottom is the whole Impulse sentence: "look
// at the top N cards of your library. Put one of them into your hand
// and the rest on the bottom of your library in any order."
//
// A LOOK, so only the looker sees the cards, and the card taken is not
// revealed — Impulse does not say so. "Put one" is mandatory: with one
// or more cards to take from, the player takes exactly one.
func LookAtTopThenTakeOneRestOnBottom(n int, label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		player := ctx.Controller()
		return TakeFromLibraryToHand{
			Player: player,
			Cards:  g.LookAtTopOfLibraryForEffect(player, n),
			Max:    1,
			Label:  label,
			Then:   TakeRestOnBottomInAnyOrder,
		}.Apply(ctx)
	}
}

// PutFromHandOnTopInAnyOrder is Brainstorm's second sentence: "put N
// cards from your hand on top of your library in any order".
//
// Two decisions and two prompts: which cards (a choose_cards over the
// hand), then in what order (a put_in_library on top). The order used
// to be the order of the picks, which is a convention in a question
// string rather than a decision the player is shown; a pile of one has
// no order, so no second prompt is raised for it.
//
// The floor is N or the whole hand, whichever is smaller: a player who
// holds fewer than N cards puts back what they have (CR 608.2 — do as
// much as you can), and an empty hand is not asked.
type PutFromHandOnTopInAnyOrder struct {
	// Player puts the cards back. Zero means the controller.
	Player uuid.UUID
	// N is how many.
	N int
	// Label is "<card> — <clause>".
	Label string
	// Then runs once the cards are on the library.
	Then func(g *game.Game) error
}

func (p PutFromHandOnTopInAnyOrder) Apply(ctx *Context) error {
	player := p.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	then := p.Then
	finish := func(g *game.Game) error {
		if then == nil {
			return nil
		}
		return then(g)
	}
	hand := allHandCardIDs(ctx.Game, player)
	if len(hand) == 0 || p.N <= 0 {
		return finish(ctx.Game)
	}
	floor := min(p.N, len(hand))
	source, label := ctx.Source(), p.Label
	order := func(g *game.Game, picked []uuid.UUID) error {
		return g.PutInLibraryInChosenOrderThenForEffect(game.PutInLibrarySpec{
			Chooser:   player,
			Source:    source,
			Cards:     picked,
			From:      game.ZoneHand,
			Placement: game.LibraryPlaceTop,
			Reason:    label + " — in what order? (the first is on top)",
			Then:      finish,
		})
	}
	if floor == len(hand) {
		// Every card in hand goes back: there is nothing to pick, only
		// an order to choose.
		return order(ctx.Game, hand)
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   source,
		Question: label,
		Cards:    hand,
		Min:      floor,
		Max:      floor,
		Zone:     game.ZoneHand,
		Then:     order,
	})
	return nil
}

// PutIntoLibrary is "put <it> into its owner's library <position>":
// on top, on the bottom, or N from the top — Oust's "second from the
// top", Teferi's and the God-Eternals' "third". No decision, so no
// prompt (ADR 0088 Decision 4). The move is the tuck route, so the CR
// 614 window opens and a commander is offered the command zone; a
// commander whose owner declines still lands at Depth, because the
// position rides the route.
type PutIntoLibrary struct {
	// Card is what moves, from wherever it is.
	Card uuid.UUID
	// Depth is the position counted from the top: 2 is "second from
	// the top". 0 and 1 are the top. A library shorter than Depth
	// takes the card on the bottom.
	Depth int
	// ToBottom is "on the bottom of its owner's library". Wins over
	// Depth.
	ToBottom bool
	// Then runs after the move, told whether the card reached a
	// library (false when a replacement or CR 903.9 sent it elsewhere).
	Then func(g *game.Game, placed bool) error
}

func (p PutIntoLibrary) Apply(ctx *Context) error {
	return ctx.Game.TuckToLibraryThenForEffect(p.Card, game.TuckOptions{
		ToBottom: p.ToBottom,
		Depth:    p.Depth,
	}, p.Then)
}

// libraryOrderLabel is "<source card name> — <clause>", or the clause
// alone when the source cannot be found.
func libraryOrderLabel(g *game.Game, source uuid.UUID, clause string) string {
	if c, ok := g.LookupCardForEffect(source); ok && c.Name != "" {
		return c.Name + " — " + clause
	}
	return clause
}
