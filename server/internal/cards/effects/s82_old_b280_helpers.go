package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b280_helpers.go — Diminishing Returns (Eternal Masters, #5542): Echo
// of Eons' wheel with a different ending. "Each player shuffles their hand and
// graveyard into their library. You exile the top ten cards of your library.
// Then each player draws up to seven cards."

// DiminishingReturnsResolve is the whole spell. The roster is TablePlayers: the
// controller first, then the opponents in turn order.
func DiminishingReturnsResolve(_ *game.StackItem, ctx *Context) error {
	players := TablePlayers(ctx)
	return diminishingReturnsShuffleEach(ctx.Game, ctx.Source(), players, players)
}

// diminishingReturnsShuffleEach tucks the head player's hand and graveyard into
// their own library as one simultaneous exit, shuffles, and continues with the
// tail — recursive for the reason echoOfEonsShuffleEach is (the tuck can pause
// on a CR 903.9b prompt). Once every player has shuffled it hands off to
// diminishingReturnsExileAndDraw. Package-level and capture-free: only IDs
// cross the boundary, in the tail's Carry (the full roster, then the seats
// still to go) with N the roster's length.
func diminishingReturnsShuffleEach(g *game.Game, source uuid.UUID, remaining, all []uuid.UUID) error {
	if len(remaining) == 0 {
		return diminishingReturnsExileAndDraw(g, source, all)
	}
	id, rest := remaining[0], remaining[1:]
	ids := append(AllHandCardIDs(g, id), EchoOfEonsGraveyardCardIDs(g, id)...)
	carry := append(append([]uuid.UUID(nil), all...), rest...)
	return g.TuckCardsToLibraryKeyedForEffect(ids, game.TuckOptions{},
		game.KeyedTail{Ref: diminishingReturnsShuffleNext, Player: id, Source: source, Carry: carry, N: len(all)})
}

// diminishingReturnsShuffleNext is one seat's shuffle and the next seat's tuck.
// An on-disk key: never renamed, never reused. Set in init, because its body
// names itself through the next tuck.
var diminishingReturnsShuffleNext game.TailRef

// RegisterDiminishingReturnsTails registers the spell's two keyed continuations;
// the card's init() calls it once (a core file declares no init()).
func RegisterDiminishingReturnsTails() {
	diminishingReturnsShuffleNext = game.RegisterTail("then/s82-old-b280-diminishing-returns-shuffle-next", func(g *game.Game, o game.TailOutcome) error {
		if err := g.ShuffleLibraryForEffect(o.Tail.Player); err != nil {
			return err
		}
		n := min(max(o.Tail.N, 0), len(o.Tail.Carry))
		all, rest := o.Tail.Carry[:n], o.Tail.Carry[n:]
		if n == 0 {
			all = nil
		}
		return diminishingReturnsShuffleEach(g, o.Tail.Source, rest, all)
	})
	registerDiminishingReturnsDraws()
}

// diminishingReturnsExileAndDraw is the second half: the controller (the
// roster's head) exiles the top ten cards of their library, then each player in
// turn is asked how many cards, up to seven, they draw.
func diminishingReturnsExileAndDraw(g *game.Game, source uuid.UUID, all []uuid.UUID) error {
	if len(all) == 0 {
		return nil
	}
	if _, err := g.MillToZoneForEffect(all[0], 10, game.ZoneExile); err != nil {
		return err
	}
	return diminishingReturnsAsk(KeyedContext(g, all[0], source), all)
}

// diminishingReturnsAsk asks the head of `left` for their number; the rest ride
// the prompt as its Carry.
func diminishingReturnsAsk(ctx *Context, left []uuid.UUID) error {
	for len(left) > 0 && ctx.Game.PlayerByIDForEffect(left[0]) == nil {
		left = left[1:]
	}
	if len(left) == 0 {
		return nil
	}
	who, rest := left[0], left[1:]
	g := ctx.Game
	return ChooseNumber{
		Chooser:  who,
		Min:      0,
		Max:      g.DrawChoiceCeilingForEffect(who, 7),
		Question: "Diminishing Returns — draw up to seven cards",
		Unit:     game.PayAmountCards,
		Goal: func(ctx *Context) int {
			if p := ctx.Game.PlayerByIDForEffect(who); p != nil && p.Library != nil && p.Library.Size() < 7 {
				return p.Library.Size()
			}
			return 7
		},
		Then:  diminishingReturnsDraws,
		Carry: rest,
	}.Apply(ctx)
}

// diminishingReturnsDraws is one player's number: they draw that many, then the
// next player is asked. An on-disk key: never renamed, never reused. Set in
// init, because its body names itself through the next question.
var diminishingReturnsDraws game.NumberThen

func registerDiminishingReturnsDraws() {
	diminishingReturnsDraws = NumberThen("number/s82-old-b280-diminishing-returns-draw-up-to-seven",
		func(ctx *Context, r game.NumberAnswer) error {
			if r.Number > 0 && ctx.Game.PlayerByIDForEffect(r.Chooser) != nil {
				if err := (DrawCards{Player: r.Chooser, N: r.Number}).Apply(ctx); err != nil {
					return err
				}
			}
			return diminishingReturnsAsk(ctx, r.Carry)
		})
}
