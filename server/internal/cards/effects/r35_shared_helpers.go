package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Helpers two card shards share, moved here from the card file that declared
// them when roll-up 35's cards (#5948) began to call them from another shard.

// R35CeruleanSphinxShuffleItself tucks the Sphinx into its owner's library,
// then that owner shuffles.
func R35CeruleanSphinxShuffleItself(g *game.Game, item *game.StackItem) error {
	if !SourceIsStillThisPermanent(g, item) {
		return nil
	}
	c, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok {
		return nil
	}
	return g.TuckCardsToLibraryKeyedForEffect([]uuid.UUID{item.SourceCardID}, game.TuckOptions{}, game.KeyedTail{Ref: ThenShuffle, Player: c.Owner})
}

// R35YouHaveNoCardsInHand is the Case's solve condition.
func R35YouHaveNoCardsInHand(g *game.Game, controller, _ uuid.UUID) bool {
	p := g.PlayerByIDForEffect(controller)
	return p != nil && p.Hand.Size() == 0
}

// R35StickerHandSize is the number of cards in `you`'s hand, walking the seats
// directly because a layer pass may hold only the read lock.
func R35StickerHandSize(g *game.Game, you uuid.UUID) int {
	for _, p := range g.Seats {
		if p != nil && p.ID == you && p.Hand != nil {
			return p.Hand.Size()
		}
	}
	return 0
}

// R35Msc18DoomsdayLoseHalfOfYourLife is shared by Doomsday; Revival // Revenge.
func R35Msc18DoomsdayLoseHalfOfYourLife(ctx *Context, who uuid.UUID) error {
	p := ctx.PlayerByID(who)
	if p == nil || p.Eliminated || p.Life <= 0 {
		return nil
	}
	return GainLife{Player: who, Amount: -((p.Life + 1) / 2)}.Apply(ctx)
}

// R35NightmareSowerCounter is shared by Nightmare Sower; Spitting Dilophosaurus.
func R35NightmareSowerCounter(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return AddCounter{Target: id, Kind: game.CounterMinusOne, N: 1}.Apply(ctx)
}

// R35EquippedCreatureOf is the creature the Equipment `source` is attached
// to now, or uuid.Nil when it is unattached or gone.
func R35EquippedCreatureOf(g *game.Game, source uuid.UUID) uuid.UUID {
	eq, ok := g.LookupCardForEffect(source)
	if !ok || !OnBattlefield(g, source) || eq.AttachedTo.Kind != game.TargetCard {
		return uuid.Nil
	}
	return eq.AttachedTo.ID
}

// R35OneLastJobReturn is shared by One Last Job; Brotherhood Outcast.
func R35OneLastJobReturn(_ *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok {
		return nil
	}
	return ReturnFromGraveyardTogether{Targets: []uuid.UUID{t.ID}, Controller: ctx.Controller()}.Apply(ctx)
}
