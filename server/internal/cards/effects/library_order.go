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
// goes in ThenTail, not on the next line, or it runs before the player
// has answered.
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

	// TopCount, when positive, is exactly how many cards go on top —
	// "put ONE of those cards on top of your library and the rest on
	// the bottom" (Cream of the Crop). Only with top-or-bottom.
	TopCount int

	// TopDepth is where the top lane lands: 2 is "second from the top"
	// (Temporal Cleansing). Zero is the top.
	TopDepth int

	// Counter makes the move a COUNTER to that position: Cards are
	// spells on the stack (Hinder, Spell Crumple). Prefer CounterToLibrary,
	// which says so.
	Counter bool

	// Label is the prompt banner, "<card> — <the printed clause>".
	Label string

	// ThenTail is the rest of the effect, run once the pile is placed:
	// a registered tail (#4301; game.RegisterTail, Context.Tail), so the
	// ordering prompt is a restore point.
	ThenTail game.KeyedTail
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
		TopCount:  p.TopCount,
		TopDepth:  p.TopDepth,
		Counter:   p.Counter,
		Reason:    p.Label,
		ThenTail:  tailDefaults(p.ThenTail, chooser, ctx.Source()),
	})
}

// LookAtLibraryThenPlace is "look at the top N cards of <a player's>
// library" followed by where the looker puts them (#1298):
//
//   - Jace, the Mind Sculptor's +2 — "look at the top card of target
//     player's library. You may put that card on the bottom of that
//     player's library": Owner = the target, N 1, top-or-bottom.
//   - Portent — "look at the top three cards of target player's
//     library, then put them back in any order": N 3, top.
//   - Cream of the Crop — "look at the top X cards of your library …
//     put one of those cards on top of your library and the rest on the
//     bottom of your library in any order": top-or-bottom, TopCount 1.
//
// A LOOK (CR 701.20): only the looker learns the cards — the library's
// owner too sees nothing it did not already know (CR 401.2). The
// placement is ADR 0088's put_in_library with the looker as chooser,
// and it puts each card back in its OWNER's library; CR 401.4 decides
// who knows the order afterwards.
//
// Like Scry it only QUEUES: anything after the placement goes in ThenTail.
type LookAtLibraryThenPlace struct {
	// Looker looks and chooses. Zero means the controller.
	Looker uuid.UUID
	// Owner is whose library. Zero means the looker's own.
	Owner uuid.UUID
	// N is how many cards from the top.
	N int
	// Placement, TopCount and TopDepth are PutInLibraryInAnyOrder's.
	Placement game.LibraryPlacement
	TopCount  int
	TopDepth  int
	// Label is "<card> — <clause>".
	Label string
	// ThenTail runs once the cards are placed (or at once, when the
	// library was empty), as PutInLibraryInAnyOrder.ThenTail (#4301).
	ThenTail game.KeyedTail
}

func (l LookAtLibraryThenPlace) Apply(ctx *Context) error {
	looker := l.Looker
	if looker == uuid.Nil {
		looker = ctx.Controller()
	}
	owner := l.Owner
	if owner == uuid.Nil {
		owner = looker
	}
	return PutInLibraryInAnyOrder{
		Chooser:   looker,
		Cards:     ctx.Game.LookAtTopOfPlayersLibraryForEffect(looker, owner, l.N),
		From:      game.ZoneLibrary,
		Placement: l.Placement,
		TopCount:  l.TopCount,
		TopDepth:  l.TopDepth,
		Label:     l.Label,
		ThenTail:  l.ThenTail,
	}.Apply(ctx)
}

// PutIntoLibraryAtDepthOrBottom is "the owner of target nonland
// permanent puts it into their library second from the top or on the
// bottom" — Temporal Cleansing, Lost Days, Wan Shi Tong, All-Knowing
// (#1298). One card, a two-way choice whose top lane is a DEPTH, and a
// chooser who is the card's OWNER, not the caster: that is the whole of
// the difference from PutIntoLibrary, which has no choice in it.
//
// The move is the tuck route either way, so the CR 614 window opens and
// a commander is offered the command zone.
type PutIntoLibraryAtDepthOrBottom struct {
	// Card is what moves.
	Card uuid.UUID
	// Depth is the top lane's position — 2 for "second from the top".
	Depth int
	// Chooser picks. Zero means the card's OWNER, which is what every
	// printed variant says.
	Chooser uuid.UUID
	// Label is "<card> — <clause>".
	Label string
}

func (p PutIntoLibraryAtDepthOrBottom) Apply(ctx *Context) error {
	c, ok := ctx.Game.LookupCardForEffect(p.Card)
	z := ctx.Game.FindCardZoneForEffect(p.Card)
	if !ok || z == nil || ctx.isNewSourceObject(p.Card) { // #1432
		return nil
	}
	chooser := p.Chooser
	if chooser == uuid.Nil {
		chooser = c.Owner
	}
	return PutInLibraryInAnyOrder{
		Chooser:   chooser,
		Cards:     []uuid.UUID{p.Card},
		From:      z.Kind,
		Placement: game.LibraryPlaceTopOrBottom,
		TopDepth:  p.Depth,
		Label:     p.Label,
	}.Apply(ctx)
}

// CounterToLibrary is "counter target spell. If that spell is
// countered this way, put it on the bottom of its owner's library"
// (Spell Crumple: LibraryPlaceBottom) and "… put that card on your
// choice of the top or bottom of its owner's library instead" (Hinder:
// LibraryPlaceTopOrBottom) — #1298.
//
// The choice comes FIRST and the counter second, because that is the
// order the card resolves in: the counterer picks the position, then
// the spell is countered to it through the shared stack exit, so CR
// 903.9 offers a commander's owner the command zone knowing where the
// card was headed, and flashback's exile still wins. A spell that can't
// be countered is not asked about and does not move; a countered
// ability ceases to exist (CR 701.6b) — use CounterTarget for a target
// that may be either.
type CounterToLibrary struct {
	// StackID is the spell.
	StackID uuid.UUID
	// Placement is LibraryPlaceBottom, LibraryPlaceTop or
	// LibraryPlaceTopOrBottom.
	Placement game.LibraryPlacement
	// Chooser picks the position. Zero means the controller — "your
	// choice".
	Chooser uuid.UUID
	// Label is "<card> — <clause>".
	Label string
}

func (c CounterToLibrary) Apply(ctx *Context) error {
	if item := ctx.Game.StackItemForEffect(c.StackID); item != nil && item.Kind != game.StackItemSpell {
		return ctx.Game.CounterTargetForEffect(c.StackID)
	}
	return PutInLibraryInAnyOrder{
		Chooser:   c.Chooser,
		Cards:     []uuid.UUID{c.StackID},
		Placement: c.Placement,
		Counter:   true,
		Label:     c.Label,
	}.Apply(ctx)
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
			Key:    takeRestBottomAny,
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
}

func (p PutFromHandOnTopInAnyOrder) Apply(ctx *Context) error {
	player := p.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	hand := allHandCardIDs(ctx.Game, player)
	if len(hand) == 0 || p.N <= 0 {
		return nil
	}
	floor := min(p.N, len(hand))
	if floor == len(hand) {
		// Every card in hand goes back: there is nothing to pick, only
		// an order to choose.
		return putFromHandOnTopInOrder(ctx.Game, player, ctx.Source(), p.Label, hand)
	}
	// Keyed (#4253): the label rides the prompt as its question, so a
	// table waiting on the pick is a restore point. The hand is
	// re-checked on submit.
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   ctx.Source(),
		Question: p.Label,
		Cards:    hand,
		Min:      floor,
		Max:      floor,
		LiveZone: game.ZoneHand,
		ThenKey:  putFromHandOnTopPicked,
	})
	return nil
}

// putFromHandOnTopPicked is the cards picked from the hand, put on top in
// the order the player chooses next. An on-disk key: never renamed, never
// reused.
var putFromHandOnTopPicked = game.RegisterChooseCardsThen("choose-cards/put-from-hand-on-top-in-any-order", func(g *game.Game, r game.CardsChosen) error {
	return putFromHandOnTopInOrder(g, r.Chooser, r.Source, r.Question, r.Picked)
})

// putFromHandOnTopInOrder is the second prompt: the order the picked cards
// go on top of the library in.
func putFromHandOnTopInOrder(g *game.Game, player, source uuid.UUID, label string, picked []uuid.UUID) error {
	return g.PutInLibraryInChosenOrderThenForEffect(game.PutInLibrarySpec{
		Chooser:   player,
		Source:    source,
		Cards:     picked,
		From:      game.ZoneHand,
		Placement: game.LibraryPlaceTop,
		Reason:    label + " — in what order? (the first is on top)",
	})
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
	// ThenTail runs after the move, a registered tail (#4253 slice 10a)
	// told the card when it reached a library and nothing when it did not
	// (a replacement or CR 903.9 sent it elsewhere; oneWentThisWay), so a
	// table waiting on the move's prompt is a restore point.
	ThenTail game.KeyedTail
}

func (p PutIntoLibrary) Apply(ctx *Context) error {
	if ctx.isNewSourceObject(p.Card) { // #1432
		return p.ThenTail.Run(ctx.Game, nil)
	}
	opts := game.TuckOptions{ToBottom: p.ToBottom, Depth: p.Depth}
	if !p.ThenTail.IsZero() {
		return ctx.Game.TuckCardsToLibraryKeyedForEffect([]uuid.UUID{p.Card}, opts, p.ThenTail)
	}
	return ctx.Game.TuckToLibraryThenForEffect(p.Card, opts, nil)
}

// libraryOrderLabel is "<source card name> — <clause>", or the clause
// alone when the source cannot be found.
func libraryOrderLabel(g *game.Game, source uuid.UUID, clause string) string {
	if c, ok := g.LookupCardForEffect(source); ok && c.Name != "" {
		return c.Name + " — " + clause
	}
	return clause
}
