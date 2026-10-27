package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Intuition — Instant {2}{U}:
//
//	"Search your library for three cards and reveal them. Target
//	 opponent chooses one. Put that card into your hand and the rest
//	 into your graveyard. Then shuffle."
//
// One of the two cards on the "opponent picks from a revealed set"
// seam row (#1214). The pieces:
//
//   - the SEARCH leaves the cards in the library (Dest: ZoneLibrary,
//     ToTop true so the search reports what it found) and does NOT
//     shuffle, because the shuffle is printed after the choice and a
//     shuffle wipes the per-card knowledge the reveal just granted;
//   - the REVEAL makes the three public, which is the only thing that
//     entitles the opponent to look at cards out of somebody else's
//     library at all (#549);
//   - the CHOICE is a reveal_pick addressed to the target, floor and
//     ceiling one, whose continuation moves both halves and then
//     shuffles.
//
// "Search for three cards" names no quality, so it is CR 701.23d's
// quantity search (#3053): the searcher must find three, or as many as
// the library holds — no Predicate, no fail to find. A library of fewer
// than three is taken whole without a prompt.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3c9faba7-f2d3-4978-be94-020dc8003dc0",
		Name:         "Intuition",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return intuitionResolve(item, ctx)
		},
	})
}

// intuitionResolve searches, then hands the three revealed cards to the
// target opponent.
//
// Caller holds g.mu.
func intuitionResolve(item *game.StackItem, ctx *Context) error {
	opp := targetedOpponent(item, ctx.Game)
	if opp == uuid.Nil {
		// CR 608.2b: the only target is gone, so the spell does
		// nothing at all — it is not "search and keep them".
		return nil
	}
	controller := ctx.Controller()
	source := ctx.Source()
	return ctx.Game.SearchLibraryThenForEffect(game.SearchLibrarySpec{
		Player:  controller,
		Source:  source,
		Dest:    game.ZoneLibrary,
		Limit:   3,
		Reveal:  true,
		Shuffle: false,
		// ToTop with Shuffle FALSE is "leave them on top of the
		// library, in the order they were found, keeping the knowledge
		// the reveal granted". Without it the search's continuation is
		// handed nothing at all — a ZoneLibrary destination with ToTop
		// unset means "leave it where it is" and reports no found
		// cards (executeSearchTakeLocked) — and the opponent would
		// have nothing to choose from. The library order it imposes is
		// wiped by the shuffle printed after the choice, so nothing
		// observable rides on it.
		ToTop:    true,
		Reason:   "Intuition — search your library for three cards",
		ThenTail: intuitionTail(ctx, opp),
	})
}

// intuitionTail is the search's continuation, the target opponent riding
// it as Carry[0]: they choose which card goes to the hand.
func intuitionTail(ctx *Context, opp uuid.UUID) game.KeyedTail {
	t := ctx.ItemTail(intuitionChoose)
	t.Carry = []uuid.UUID{opp}
	return t
}

// intuitionChoose is the continuation of the search: the item and the
// opponent are data on the tail, so an undo or a restore across either
// prompt resolves it against the restored game. An on-disk key.
var intuitionChoose = CardTail("intuition/choose", func(ctx *Context, o game.TailOutcome) error {
	found := o.IDs
	if len(found) == 0 {
		// Nothing was found, so there is nothing to choose from
		// and nothing to move. The shuffle still happens.
		return ctx.Game.ShuffleLibraryForEffect(ctx.Controller())
	}
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	return RevealPick{
		Player:   o.Tail.Carry[0],
		Owner:    ctx.Controller(),
		Question: "Intuition — choose one of these cards; it goes to their hand and the rest to their graveyard",
		Cards:    found,
		Min:      1,
		Max:      1,
		Then:     intuitionSettle,
	}.Apply(ctx)
})

// intuitionSettle moves the chosen card to hand, the rest to the
// graveyard, and shuffles. A card that is no longer where the search
// left it is skipped rather than erroring — the prompt is
// asynchronous and the library can move under it.
//
// Caller holds g.mu.
func intuitionSettle(ctx *Context, picked, left []uuid.UUID) error {
	return revealPickSettle(ctx, picked, left, true)
}

// revealPickSettle is the body Intuition and Gifts Ungiven share: one
// half to hand, the other to the graveyard, then shuffle. `toHandFirst`
// says which half the chooser's pick is — Intuition's pick goes to the
// hand, Gifts Ungiven's to the graveyard, and that inversion is the
// whole difference between the two cards' last sentences.
//
// The shuffle is hung off the BOUNCE's continuation rather than
// written after it (ADR 0013 §5t / §5v): a card returning to a hand is
// a zone change, the CR 614 window can pause it, and a clause written
// after a fire-and-forget exit is a payout on a move that has not
// happened yet. The graveyard half goes first precisely so the bounce
// is the LAST exit and can carry the rest of the card.
//
// A card that is no longer where the search left it is skipped rather
// than erroring — the prompt is asynchronous and the library can move
// under it.
//
// Caller holds g.mu.
func revealPickSettle(ctx *Context, picked, left []uuid.UUID, toHandFirst bool) error {
	toHand, toGraveyard := picked, left
	if !toHandFirst {
		toHand, toGraveyard = left, picked
	}
	for _, id := range toGraveyard {
		if ctx.Game.FindCardZoneForEffect(id) == nil {
			continue
		}
		if err := ctx.Game.PutIntoGraveyardForEffect(id); err != nil {
			return err
		}
	}
	// The shuffle is the move's keyed tail (#4253 slice 10a).
	return ctx.Game.BounceCardsToHandKeyedForEffect(liveCards(ctx.Game, toHand), ctx.Tail(thenShuffle))
}

// liveCards drops the IDs that are no longer in any zone, so a batch
// exit is asked only about cards that are still somewhere.
//
// Caller holds g.mu.
func liveCards(g *game.Game, ids []uuid.UUID) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if g.FindCardZoneForEffect(id) != nil {
			out = append(out, id)
		}
	}
	return out
}

// targetedOpponent reads a spell's single player target, or uuid.Nil
// when it has none or the seat has left (CR 608.2b).
//
// Caller holds g.mu.
func targetedOpponent(item *game.StackItem, g *game.Game) uuid.UUID {
	if item == nil || len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
		return uuid.Nil
	}
	id := item.Targets[0].ID
	if p := g.PlayerByIDForEffect(id); p == nil || p.Eliminated {
		return uuid.Nil
	}
	return id
}
