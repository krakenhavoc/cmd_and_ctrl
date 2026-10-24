package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// each_player_in_turn.go: "each player mills ...", "exile the top card of
// each opponent's library" and the other multi-library instructions, done
// one library at a time (#4266).
//
// Why a chain and not a loop. A mill or an exile can stop to ask a question:
// a CR 616.1 ordering prompt between two replacement effects, a commander's
// CR 903.9 offer. A loop that calls the fire-and-forget form once per player
// starts every player's mill on one line, so every library that has something
// to ask opens its prompt at once and the table answers them in whatever order
// the UI shows them. Players who must choose at the same time do it in APNAP
// order (CR 101.4, and CR 616.1 for the replacement choices a mill asks): the
// active player first, then each other player in turn order. A chain makes
// that real. Each player's mill starts from the previous player's continuation, so at
// most one of these prompts is open at a time, the active player's first, and
// the next library is not touched until it is answered. Dread Summons
// (b35DreadSummonsMillStep) and Doomsday Excruciator
// (doomsdayExileFrom) do the same by hand; a card that does it for several
// players calls the helpers below instead of writing a third copy.

// PlayerCards is one player's share of a multi-player instruction: whose
// library it was done to and which cards reached the destination zone (CR
// 400.7, the cards that ARRIVED), top of the library first.
type PlayerCards struct {
	Player uuid.UUID
	Cards  []uuid.UUID
}

// apnapOrder returns the members of `players` that are still in the game, in
// APNAP order (CR 101.4). A player named twice is listed once.
func apnapOrder(g *game.Game, players []uuid.UUID) []uuid.UUID {
	want := make(map[uuid.UUID]bool, len(players))
	for _, id := range players {
		want[id] = true
	}
	var out []uuid.UUID
	for _, id := range apnapPlayers(g) {
		if want[id] {
			out = append(out, id)
		}
	}
	return out
}

// eachPlayerInTurn runs `act` for each of `players` in APNAP order, each one
// from the previous one's continuation, and then runs `done` with every
// player's result in that order. `act` must call `next` exactly once, with the
// cards it did the instruction to, from wherever its own instruction ends:
// inline on a clean run, from the answer when a prompt paused it. `done` is
// where the card does anything that follows the instruction (a count, a cast
// permission, a reanimation), because that waits for the last player; it may
// be nil.
//
// The results are carried forward BY VALUE, never through a variable the
// continuations share: an undo across a prompt replays the continuation, and a
// list built up in a captured slice would count the first run's cards twice
// (the property b35DreadSummonsMillStep documents).
func eachPlayerInTurn(
	g *game.Game,
	players []uuid.UUID,
	act func(g *game.Game, player uuid.UUID, next func(g *game.Game, cards []uuid.UUID) error) error,
	done func(g *game.Game, results []PlayerCards) error,
) error {
	return eachPlayerFrom(g, apnapOrder(g, players), nil, act, done)
}

func eachPlayerFrom(
	g *game.Game,
	players []uuid.UUID,
	results []PlayerCards,
	act func(g *game.Game, player uuid.UUID, next func(g *game.Game, cards []uuid.UUID) error) error,
	done func(g *game.Game, results []PlayerCards) error,
) error {
	if len(players) == 0 {
		if done == nil {
			return nil
		}
		return done(g, results)
	}
	head, rest := players[0], players[1:]
	return act(g, head, func(g *game.Game, cards []uuid.UUID) error {
		// A fresh slice per step: the previous step's list is not ours to
		// append into, because an undo replays from it.
		so := make([]PlayerCards, len(results), len(results)+1)
		copy(so, results)
		so = append(so, PlayerCards{Player: head, Cards: append([]uuid.UUID(nil), cards...)})
		return eachPlayerFrom(g, rest, so, act, done)
	})
}

// millEachPlayerInTurn is "each of these players mills N cards": each library
// in APNAP order, one prompt at a time, then `done` with what each milled.
// `n` is asked once per player BEFORE any of them mills, so a count that reads
// the table ("cards equal to the number of cards in their hand", Folio of
// Fancies) is read as the instruction begins, as the paper instruction does,
// and not after an earlier player's mill. A player whose N is not positive
// mills nothing and is still in the results, with no cards.
//
// `ctx` is only used for its item and game; `done` gets a context rebuilt from
// the live game, as MillToZone.Then does.
func millEachPlayerInTurn(
	ctx *Context,
	players []uuid.UUID,
	n func(player uuid.UUID) int,
	done func(ctx *Context, results []PlayerCards) error,
) error {
	counts := make(map[uuid.UUID]int, len(players))
	for _, id := range players {
		counts[id] = n(id)
	}
	item := ctx.Item
	return eachPlayerInTurn(ctx.Game, players,
		func(g *game.Game, player uuid.UUID, next func(g *game.Game, cards []uuid.UUID) error) error {
			return MillToZone{
				Player: player,
				N:      counts[player],
				Then: func(ctx *Context, milled []uuid.UUID) error {
					return next(ctx.Game, milled)
				},
			}.Apply(NewContext(g, item))
		},
		contextDone(item, done))
}

// exileTopEachPlayerInTurn is "exile the top card of each of these players'
// libraries, and <grant>": template names the grant (its From is overwritten
// with each player in turn), and `done` gets the cards that reached exile from
// each library. Face up and unbudgeted only: see ExileTopWithPermission.ApplyThen.
func exileTopEachPlayerInTurn(
	ctx *Context,
	players []uuid.UUID,
	template ExileTopWithPermission,
	done func(ctx *Context, results []PlayerCards) error,
) error {
	item := ctx.Item
	return eachPlayerInTurn(ctx.Game, players,
		func(g *game.Game, player uuid.UUID, next func(g *game.Game, cards []uuid.UUID) error) error {
			e := template
			e.From = player
			return e.ApplyThen(NewContext(g, item), func(ctx *Context, landed []uuid.UUID) error {
				return next(ctx.Game, landed)
			})
		},
		contextDone(item, done))
}

// contextDone adapts a card's continuation, written over a Context, to
// eachPlayerInTurn's, which is handed the live game.
func contextDone(item *game.StackItem, done func(ctx *Context, results []PlayerCards) error) func(g *game.Game, results []PlayerCards) error {
	if done == nil {
		return nil
	}
	return func(g *game.Game, results []PlayerCards) error {
		return done(NewContext(g, item), results)
	}
}

// constantMill is the N of a mill that names one number for every player.
func constantMill(n int) func(uuid.UUID) int {
	return func(uuid.UUID) int { return n }
}
