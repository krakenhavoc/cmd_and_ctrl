package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b38_helpers.go — shared pieces of Avatar: The Last Airbender
// Eternal, batches 38 to 40 (#5075). Append-only.

// avB38LandCreaturesYouControl is "land creature you control" read
// through the layers: a land an earthbend animated counts while it is a
// creature.
func avB38LandCreaturesYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.Controller == source.Controller && target.IsLand() && target.IsCreature()
}

// avB38TwoCountersOnEachLandCreatureYouControl is Bumi, Eclectic
// Earthbender's attack trigger: the land creatures the controller
// controls as it resolves, snapshotted first, each get two +1/+1
// counters as one placement (a CR 616 prompt over any of them is a
// restore point).
func avB38TwoCountersOnEachLandCreatureYouControl(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	g.RecomputeLayersIfStaleLocked()
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsLand() && c.IsCreature() {
			ids = append(ids, c.InstanceID)
		}
	}
	return avB24PutCountersOnEachThen(ctx, ids, game.CounterPlusOne, 2, game.KeyedTail{})
}
