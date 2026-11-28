package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b734_helpers.go — S82 OLD batch 734 (Portal Second Age, #5625).

// S82OldB734ShuffleDiedCardIntoLibrary is "shuffle it into its owner's
// library" in a dies trigger (Angel of Fury): the card goes into its owner's
// library if it is still in the graveyard as the card that died (CR 400.7),
// and that library is shuffled whether or not it did (CR 701.24c).
func S82OldB734ShuffleDiedCardIntoLibrary(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	id := item.Trigger.Event.CardID
	owner := item.Controller
	if c, ok := g.LookupCardForEffect(id); ok && c.Owner != uuid.Nil {
		owner = c.Owner
	}
	if !DiedCardStillInGraveyard(NewContext(g, item), id) {
		return g.ShuffleLibraryForEffect(owner)
	}
	return g.TuckCardsToLibraryKeyedForEffect([]uuid.UUID{id}, game.TuckOptions{}, game.KeyedTail{Ref: ThenShuffle, Player: owner})
}

// S82OldB734DiscardTargetPlayer is "have target opponent discard a card": the
// first still-legal player target discards a card of their choice, and a
// target that left leaves nothing to do (CR 608.2b).
func S82OldB734DiscardTargetPlayer(g *game.Game, item *game.StackItem) error {
	for _, t := range NewContext(g, item).LegalTargets() {
		if t.Kind == game.TargetPlayer && t.ID != uuid.Nil {
			return DiscardOne(g, item, t.ID)
		}
	}
	return nil
}

// S82OldB734SacrificeUnlessDiscardACreature is "sacrifice this creature
// unless you discard a creature card" (Hidden Horror): a player with no
// creature card in hand sacrifices it at once; otherwise they pick a creature
// card to discard, or none, which sacrifices it if it is still the same
// permanent.
func S82OldB734SacrificeUnlessDiscardACreature(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if !SourceIsStillThisPermanent(g, item) {
			return nil
		}
		p := g.PlayerByIDForEffect(item.Controller)
		if p == nil || p.Eliminated {
			return nil
		}
		hasCreature := false
		if p.Hand != nil {
			for _, c := range p.Hand.Cards {
				if c.IsCreature() {
					hasCreature = true
				}
			}
		}
		if !hasCreature {
			return SacrificePermanent{Target: item.SourceCardID}.Apply(NewContext(g, item))
		}
		return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
			Player:   item.Controller,
			Source:   item.SourceCardID,
			N:        1,
			UpTo:     true,
			Question: name + " — discard a creature card, or sacrifice it?",
			Asked:    game.NewAskedItem(item),
		}, s82OldB734Discarded, nil)
	}
}

// s82OldB734Discarded is the discard answered: only a creature card is a
// legal pick, and no pick at all sacrifices the permanent. An on-disk key:
// never renamed, never reused.
var s82OldB734Discarded = game.RegisterDiscardThenRule("choose-cards/s82-old-b734-sacrifice-unless-discard-creature",
	func(picked []game.Card) bool { return len(picked) == 1 && picked[0].IsCreature() },
	func(g *game.Game, r game.DiscardsDone) error {
		if r.Item == nil || len(r.Discarded.Cards()) > 0 {
			return nil
		}
		if !SourceIsStillThisPermanent(g, r.Item) {
			return nil
		}
		return SacrificePermanent{Target: r.Item.SourceCardID}.Apply(NewContext(g, r.Item))
	})
