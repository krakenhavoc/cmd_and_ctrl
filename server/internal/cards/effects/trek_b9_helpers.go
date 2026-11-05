package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b9_helpers.go — Star Trek, batch 9 (#4973): the pieces the batch's
// cards share.

// trekB9DifferentlyNamedPermanentsYouControl is the cost-side count of
// "the number of differently named <matching permanents> you control":
// each matching permanent the activator controls, counted once per English
// name, read off the live battlefield as the cost is determined (CR 602.2b,
// 601.2f). Suus Mahna Practitioner's strange new worlds passes Land().
func trekB9DifferentlyNamedPermanentsYouControl(match CardPredicate) func(q game.CostQuery) int {
	return func(q game.CostQuery) int {
		if q.Game == nil || q.Game.Battlefield == nil {
			return 0
		}
		names := map[string]bool{}
		for _, c := range q.Game.Battlefield.Cards {
			if c.Controller == q.Controller && match(q.Game, q.Controller, c) {
				names[c.Name] = true
			}
		}
		return len(names)
	}
}

// trekB9TapAndStunTheModesTarget is "Tap up to one target artifact or
// creature. Put a stun counter on it." as a modal bullet's body (Subcommander
// T'Pol): the bullet's target, if it chose one and it is still legal, is
// tapped and gets one stun counter (CR 122.1d).
func trekB9TapAndStunTheModesTarget(_ *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	if err := (TapTarget{Target: t.ID}).Apply(ctx); err != nil {
		return err
	}
	return AddCounter{Target: t.ID, Kind: game.CounterStun, N: 1}.Apply(ctx)
}

// trekB9ShuffleTargetsIntoOwnersLibraries is "The owners of those cards
// shuffle them into their libraries" over every still-legal target card in a
// graveyard (Those Old Scientists, chapter III): the cards go to their
// owners' libraries in one move (a library is a CR 903.9 destination, so a
// commander's owner is asked), then each of those owners shuffles. CR
// 701.24c: the owners shuffle whether or not their card arrived.
func trekB9ShuffleTargetsIntoOwnersLibraries(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids, owners []uuid.UUID
	seen := map[uuid.UUID]bool{}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		c, ok := g.LookupCardForEffect(t.ID)
		if !ok {
			continue
		}
		ids = append(ids, t.ID)
		if !seen[c.Owner] {
			seen[c.Owner] = true
			owners = append(owners, c.Owner)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return g.TuckCardsToLibraryKeyedForEffect(ids, game.TuckOptions{},
		game.KeyedTail{Ref: thenShuffleEachCarried, Player: item.Controller, Source: item.SourceCardID, Carry: owners})
}
