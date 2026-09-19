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
// One mana, three cards seen, net zero. The batch-01 triage filed it
// under "library top", and the two halves it wanted both exist: the
// `choose_cards` prompt over a hand (with the Zone re-check that
// makes a pick still-in-hand at submit) and
// `game.TuckToLibraryForEffect`, which moves a card to the top of its
// owner's library FROM WHEREVER IT IS — the hand included.
//
// "Then" is load-bearing and is why the prompt is raised from inside
// the draw's continuation rather than on the next line: the two cards
// put back are chosen out of the hand the draw left, which is what
// makes Brainstorm a shuffle-effect combo piece instead of a cantrip.
//
// # The order
//
// The cards go back in the order they are picked: the FIRST pick ends
// on top, so a player who wants to draw a card back next turn picks
// it first. "In any order" is the player's choice and this is how the
// prompt spells it; there is no second prompt for the ordering.
//
// # A short hand
//
// The floor is two or the whole hand, whichever is smaller. A player
// who drew into an empty library and holds one card puts that card
// back rather than owing an answer nobody can give (CR 608.2 — do as
// much as you can), and a player with an empty hand is not asked.
//
// A commander among the two gets the CR 903.9 offer on the way to the
// library, because the tuck routes through the shared exit primitive;
// its owner may send it to the command zone instead, and Brainstorm
// has nothing left to do either way.
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
			return putTwoFromHandOnTopOfLibrary(ctx, item.Controller, "Brainstorm")
		},
	})
}

// putTwoFromHandOnTopOfLibrary is Brainstorm's second sentence: pick
// two cards out of `player`'s hand and put them on top of their
// library, first pick on top.
//
// The picks are tucked in REVERSE order because each tuck pushes to
// the top of the library, so the last one moved is the one that ends
// up there.
func putTwoFromHandOnTopOfLibrary(ctx *Context, player uuid.UUID, cardName string) error {
	hand := allHandCardIDs(ctx.Game, player)
	if len(hand) == 0 {
		return nil
	}
	floor := 2
	if len(hand) < floor {
		floor = len(hand)
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   ctx.Source(),
		Question: cardName + " — put two cards from your hand on top of your library (first pick on top)",
		Cards:    hand,
		Min:      floor,
		Max:      floor,
		Zone:     game.ZoneHand,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			for i := len(picked) - 1; i >= 0; i-- {
				if err := g.TuckToLibraryForEffect(picked[i], false); err != nil {
					return err
				}
			}
			return nil
		},
	})
	return nil
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
