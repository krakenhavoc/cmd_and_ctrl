package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Loyal Sentry — Creature — Human Soldier {W}, 1/1 (Starter 1999,
// #5620):
//
//	"When this creature blocks a creature, destroy that creature and this
//	 creature."
//
// Labyrinth Minotaur's per-pair block trigger (CR 509.3b): EventBlock
// names the Sentry in CardID and the attacker in Target, one event per
// pair of the final declaration. The trigger stamps the attacker as an
// object (Params.Object, CR 400.7), so an attacker that left the
// battlefield and came back before it resolved is a new object and is
// not destroyed; a Sentry that did the same is not destroyed either.
// Both are destroyed at the same time (Dead-Iron Sledge's
// DestroyPermanentsForEffect), so either may regenerate or be
// indestructible, and one that is already gone leaves the other to be
// destroyed alone.
//
// No simplification.
func init() {
	const label = "Loyal Sentry — destroy the creature it blocks and this creature"
	t := On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return b28SelfBlocked(ev, source)
	}, label, loyalSentryDestroyBoth)
	t.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Object = game.ObjectRef{ID: ev.Target, Epoch: g.ObjectEpochForEffect(ev.Target)}
		return item
	}
	Register(Spec{
		OracleID:     "54b1a496-1f8e-4581-8260-bad57b25c6e3",
		Name:         "Loyal Sentry",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}

// loyalSentryDestroyBoth destroys the blocked attacker, if it is still
// the object the Sentry blocked, and the Sentry, if it is still the
// object whose ability this is, in one simultaneous destruction.
func loyalSentryDestroyBoth(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	attacker := item.Params.Object
	if attacker.ID != uuid.Nil && onBattlefield(g, attacker.ID) && g.ObjectEpochForEffect(attacker.ID) == attacker.Epoch {
		ids = append(ids, attacker.ID)
	}
	if onBattlefield(g, item.SourceCardID) && !ctx.isNewSourceObject(item.SourceCardID) {
		ids = append(ids, item.SourceCardID)
	}
	if len(ids) > 0 {
		g.DestroyPermanentsForEffect(ids)
	}
	return nil
}
