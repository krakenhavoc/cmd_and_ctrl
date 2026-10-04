package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// counters_on_creatures.go — shared bodies that put a named counter on
// creatures (#1858, Bomb Squad's fuse counters). Append-only.

// putACounterOnTheTarget is "Put a <kind> counter on target creature":
// one counter on the first target, if it is still legal as the ability
// resolves (CR 608.2b).
func putACounterOnTheTarget(kind string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		if len(item.Targets) == 0 || !ctx.IsTargetLegal(item.Targets[0]) {
			return nil
		}
		return AddCounter{Target: item.Targets[0].ID, Kind: kind, N: 1}.Apply(ctx)
	}
}

// putACounterOnEachCreatureWith is "put a <kind> counter on each
// creature with a <kind> counter on it", whoever controls it. The set is
// read before the first counter goes on, so a creature is counted once.
func putACounterOnEachCreatureWith(kind string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		var ids []uuid.UUID
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.IsCreature() && c.Counters[kind] > 0 {
				ids = append(ids, c.InstanceID)
			}
		}
		ctx := NewContext(g, item).asGroupMember()
		for _, id := range ids {
			if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
				continue
			}
			if err := (AddCounter{Target: id, Kind: kind, N: 1}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	}
}

// putAPlusOneCounterOnEachYouControl is "put a +1/+1 counter on each
// <noun> you control" — Nazgûl's "each Wraith you control": every
// permanent the resolving item's controller controls that `match`
// admits. The set is read before the first counter goes on, and each
// counter settles before the next (b13PutCounterOnEach).
func putAPlusOneCounterOnEachYouControl(match CardPredicate) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		var ids []uuid.UUID
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller == item.Controller && match(g, item.Controller, c) {
				ids = append(ids, c.InstanceID)
			}
		}
		return b13PutCounterOnEach(NewContext(g, item), ids)
	}
}
