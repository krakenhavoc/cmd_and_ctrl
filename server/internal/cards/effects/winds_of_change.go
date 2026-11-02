package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Winds of Change — Sorcery {R} (Marvel's Spider-Man, #4460):
//
//	"Each player shuffles the cards from their hand into their library,
//	 then draws that many cards."
//
// Echo of Eons's two-phase shape without the graveyard: each player's
// hand goes into their own library as one simultaneous exit
// (TuckCardsToLibraryKeyedForEffect, so a commander in a hand is offered
// the command zone, CR 903.9b), then that library is shuffled, seat by
// seat, and only once every seat has shuffled does anyone draw. "That
// many" is the number of that player's cards that actually reached the
// library: a commander that went to the command zone instead was not
// shuffled in and is not counted.
//
// The roster rides the tail's Carry and the counts so far its Nums, so a
// table paused on a commander's question is a restore point.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f525cf10-e24c-4c46-9a13-6f8579d09d50",
		Name:         "Winds of Change",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return windsOfChangeShuffleFrom(ctx.Game, tablePlayers(ctx), nil)
		},
	})
}

// windsOfChangeShuffleFrom tucks the hand of the seat after the ones
// already counted (len(counts) of them), or, once every seat is done,
// deals each its new hand.
func windsOfChangeShuffleFrom(g *game.Game, all []uuid.UUID, counts []int) error {
	if len(counts) >= len(all) {
		for i, id := range all {
			if counts[i] <= 0 {
				continue
			}
			if err := g.DrawNForEffect(id, counts[i]); err != nil {
				return err
			}
		}
		return nil
	}
	id := all[len(counts)]
	return g.TuckCardsToLibraryKeyedForEffect(allHandCardIDs(g, id), game.TuckOptions{},
		game.KeyedTail{Ref: windsOfChangeShuffled, Player: id, Carry: append([]uuid.UUID(nil), all...),
			Nums: append([]int(nil), counts...)})
}

// windsOfChangeShuffled is one seat's shuffle: the cards that reached the
// library are counted, the library is shuffled, and the next seat goes.
// An on-disk key: never renamed, never reused.
var windsOfChangeShuffled game.TailRef

func init() {
	windsOfChangeShuffled = game.RegisterTail("then/winds-of-change-shuffled", func(g *game.Game, o game.TailOutcome) error {
		if err := g.ShuffleLibraryForEffect(o.Tail.Player); err != nil {
			return err
		}
		counts := append(append([]int(nil), o.Tail.Nums...), len(o.IDs))
		return windsOfChangeShuffleFrom(g, o.Tail.Carry, counts)
	})
}
