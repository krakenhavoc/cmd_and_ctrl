package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// effects_counter_costs.go — effect bodies for the abilities #625's
// counter-removal cost made expressible (Dragon's Hoard, Mikaeus, the
// Lunarch, Benevolent Hydra, Fain, the Broker). Append-only, like every
// shared vocabulary file.

// drawOneCard is "Draw a card." for the activator.
func drawOneCard(g *game.Game, item *game.StackItem) error {
	return g.DrawNForEffect(item.Controller, 1)
}

// createOneTreasure is "Create a Treasure token." for the activator.
func createOneTreasure(g *game.Game, item *game.StackItem) error {
	return CreateToken{Controller: item.Controller, Template: TreasureToken(), N: 1}.Apply(NewContext(g, item))
}

// putCounterOnEachOtherCreatureYouControl is Mikaeus, the Lunarch's
// "Put a +1/+1 counter on each other creature you control." "Other" is
// the ability's source, read off the item rather than captured. The set
// is snapshotted before the first counter lands, so a creature that a
// counter payoff makes mid-loop gets nothing.
func putCounterOnEachOtherCreatureYouControl(g *game.Game, item *game.StackItem) error {
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsCreature() && c.InstanceID != item.SourceCardID {
			ids = append(ids, c.InstanceID)
		}
	}
	return b13PutCounterOnEach(NewContext(g, item), ids)
}

// putCounterOnChosenCreature puts one +1/+1 counter on the ability's
// target if it is still legal — Benevolent Hydra's "Put a +1/+1 counter
// on another target creature you control."
func putCounterOnChosenCreature(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
	}
	return nil
}
