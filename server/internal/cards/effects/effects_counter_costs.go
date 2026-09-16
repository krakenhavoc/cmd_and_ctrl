package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// effects_counter_costs.go — effect bodies for the abilities #625's
// counter-removal cost made expressible that no existing helper covers
// (Dragon's Hoard, Benevolent Hydra and Fain, the Broker reuse
// b27DrawOne, b36CounterOnChosenAnimal and the house Treasure one-liner).
// Append-only, like every shared vocabulary file.

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
