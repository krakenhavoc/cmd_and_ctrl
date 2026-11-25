package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82OldB121CounterOnEachOtherOfType is "put a +1/+1 counter on each other
// <subtype> you control" (Thalia's Lieutenant): every other permanent the
// trigger's controller controls with the subtype in its effective view, so a
// changeling counts. The set is read as the trigger resolves (CR 608.2c).
func S82OldB121CounterOnEachOtherOfType(subtype string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		var ids []uuid.UUID
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller == item.Controller && c.InstanceID != item.SourceCardID && c.IsCreature() && c.HasSubtype(subtype) {
				ids = append(ids, c.InstanceID)
			}
		}
		return B13PutCounterOnEach(NewContext(g, item), ids)
	}
}
