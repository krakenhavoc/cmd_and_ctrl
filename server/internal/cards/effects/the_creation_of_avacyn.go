package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Creation of Avacyn — Enchantment — Saga {1}{B}{B}:
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Search your library for a card, exile it face down, then
//	     shuffle.
//	 II — Turn the exiled card face up. If it's a creature card, you
//	     lose life equal to its mana value.
//	 III — You may put the exiled card onto the battlefield if it's a
//	     creature card. If you don't put it onto the battlefield, put
//	     it into its owner's hand."
//
// "The exiled card" is CR 607.2a's linked-ability reference: chapters II
// and III mean the card chapter I exiled, and nothing else. Chapter I
// therefore links what it finds to this Saga OBJECT (SearchLibrary.LinkTo,
// FaceDownHidden), and the later chapters read it back with
// ExiledWithForEffect on the same {instance, epoch} pair. A Saga that was
// bounced and replayed is a new object with no claim on the card, and the
// card then stays face down in exile, which is the printed behaviour.
//
// FaceDownHidden is the viewer rule this card prints: a face-down exile
// that nobody else may look at, and its controller (the player who
// searched) already saw it when they picked it (CR 406.3). Chapter II turns
// it face up in exile (game.TurnExiledFaceUpForEffect), which shows it to
// every seat and announces it in the log.
//
// The search has no stated quality, so it must find a card when the library
// has one (CR 701.23d) and the searcher picks which. An empty library finds
// nothing and the later chapters do nothing. The life loss reads the card's
// mana value as it is face up in exile (CR 202.3); a non-creature card
// costs nothing. Chapter III's "you may" is a yes/no prompt only for a
// creature card; anything else goes straight to its owner's hand.
func init() {
	const name = "The Creation of Avacyn"
	Register(Spec{
		OracleID:     "7f32344f-06d8-4862-bfeb-f90a3e1d7fc2",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, "search for a card, exile it face down"), avacynSearch),
			ChapterTrigger(2, SagaChapterLabel(name, 2, "turn the exiled card face up"), avacynTurnUp),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "put the exiled card onto the battlefield or into hand"), avacynFinish),
		},
	})
}

func avacynSearch(g *game.Game, item *game.StackItem) error {
	ref, ok := avacynSaga(g, item)
	if !ok {
		return nil
	}
	return SearchLibrary{
		Player:   item.Controller,
		Dest:     game.ZoneExile,
		Limit:    1,
		FaceDown: game.FaceDownHidden,
		LinkTo:   ref,
		Shuffle:  true,
		Reason:   "The Creation of Avacyn — search your library for a card",
	}.Apply(NewContext(g, item))
}

func avacynTurnUp(g *game.Game, item *game.StackItem) error {
	ref, ok := avacynSaga(g, item)
	if !ok {
		return nil
	}
	for _, id := range avacynExiled(g, ref) {
		if !g.TurnExiledFaceUpForEffect(item.SourceCardID, id, "The Creation of Avacyn — turn the exiled card face up") {
			continue
		}
		c, found := g.LookupCardForEffect(id)
		if !found || !c.IsCreature() {
			continue
		}
		if mv := c.ManaValue(); mv > 0 {
			if err := g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -mv); err != nil {
				return err
			}
		}
	}
	return nil
}

func avacynFinish(g *game.Game, item *game.StackItem) error {
	ref, ok := avacynSaga(g, item)
	if !ok {
		return nil
	}
	for _, id := range avacynExiled(g, ref) {
		if err := g.PutExiledCreatureOntoBattlefieldOrHandForEffect(item.Controller, item.SourceCardID, id); err != nil {
			return err
		}
	}
	return nil
}

// avacynSaga is the Saga object this chapter belongs to, as the chapter
// trigger captured it. False for an item with no source object.
func avacynSaga(g *game.Game, item *game.StackItem) (game.PermissionCardRef, bool) {
	ref, ok := g.SourceObjectForEffect(item)
	return game.PermissionCardRef(ref), ok
}

// avacynExiled lists the instance IDs of the cards in exile linked to the
// Saga object `ref`, in exile order. Copied out of the exile zone so the
// chapter may move them.
func avacynExiled(g *game.Game, ref game.PermissionCardRef) []uuid.UUID {
	var ids []uuid.UUID
	for _, c := range g.ExiledWithForEffect(ref) {
		ids = append(ids, c.InstanceID)
	}
	return ids
}
