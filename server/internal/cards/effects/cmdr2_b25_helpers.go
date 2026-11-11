package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_b25_helpers.go — shared bodies for the -1/-1 counter placers of
// Commander wave 2, batch 25 (#5368). Append-only.

// cmdr2B25CountersOnTarget is "put N -1/-1 counters on target creature":
// the first target, if it is still legal as the ability resolves
// (CR 608.2b).
func cmdr2B25CountersOnTarget(n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		if len(item.Targets) == 0 || !ctx.IsTargetLegal(item.Targets[0]) {
			return nil
		}
		return AddCounter{Target: item.Targets[0].ID, Kind: game.CounterMinusOne, N: n}.Apply(ctx)
	}
}

// cmdr2B25CountersOnEach is "put a -1/-1 counter on each creature" that
// `match` admits (nil admits every creature), except the source when
// others is set ("each other creature"). The set is read before the first
// counter goes on, so a creature is counted once, and each is a member of
// a group (CR 608.2h).
func cmdr2B25CountersOnEach(others bool, match CardPredicate) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		var ids []uuid.UUID
		for _, c := range g.BattlefieldCardsForEffect() {
			if !c.IsCreature() || (others && c.InstanceID == item.SourceCardID) {
				continue
			}
			if match != nil && !match(g, item.Controller, c) {
				continue
			}
			ids = append(ids, c.InstanceID)
		}
		for _, id := range ids {
			if !onBattlefield(g, id) {
				continue
			}
			if err := (AddCounter{Target: id, Kind: game.CounterMinusOne, N: 1}).Apply(ctx.asGroupMember()); err != nil {
				return err
			}
		}
		return nil
	}
}
