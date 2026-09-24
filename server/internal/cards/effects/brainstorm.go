package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Brainstorm — Instant {U}:
//
//	"Draw three cards, then put two cards from your hand on top of
//	 your library in any order."
//
// One mana, three cards seen, net zero. The two halves are the draw
// and PutFromHandOnTopInAnyOrder (library_order.go, ADR 0088).
//
// "Then" is load-bearing and is why the put-back is raised after the
// draw rather than beside it: the two cards put back are chosen out of
// the hand the draw left, which is what makes Brainstorm a
// shuffle-effect combo piece instead of a cantrip.
//
// # The order
//
// "In any order" is a second decision, and since #996 it is a second
// prompt: pick the two cards, then arrange them (a put_in_library on
// top — the first card of the answer is the next draw). Until then the
// order was the order of the picks, a convention that lived in the
// question string and that the client's card grid never showed.
//
// # A short hand
//
// The floor is two or the whole hand, whichever is smaller. A player
// who drew into an empty library and holds one card puts that card
// back rather than owing an answer nobody can give (CR 608.2 — do as
// much as you can), and a player with an empty hand is not asked.
//
// A commander among the two gets the CR 903.9 offer on the way to the
// library, because the placement routes through the shared exit
// primitive; its owner may send it to the command zone instead.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "36cd2364-d113-47d1-b2c4-b088d9eb88dd",
		Name:         "Brainstorm",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (DrawCards{Player: item.Controller, N: 3}.Apply(ctx)); err != nil {
				return err
			}
			return PutFromHandOnTopInAnyOrder{
				Player: item.Controller,
				N:      2,
				Label:  "Brainstorm — put two cards from your hand on top of your library",
			}.Apply(ctx)
		},
	})
}

// allHandCardIDs is every card in a player's hand, in hand order.
// Unlike handCardsMatching it keeps nonpermanent cards, because a
// clause that puts cards BACK does not care what they could become.
func allHandCardIDs(g *game.Game, player uuid.UUID) []uuid.UUID {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Hand == nil {
		return nil
	}
	out := make([]uuid.UUID, 0, len(p.Hand.Cards))
	for _, c := range p.Hand.Cards {
		out = append(out, c.InstanceID)
	}
	return out
}
