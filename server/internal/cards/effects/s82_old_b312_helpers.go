package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b312_helpers.go — Battle for Zendikar pump, combat and counter
// cards (#5547).

// S82OldB312IsolationCounter is the isolation counter (Quarantine Field).
// Counter names are free-form strings, so no registry entry is needed.
const S82OldB312IsolationCounter = "isolation"

// S82OldB312PlusPerOtherAttacking is "Whenever this creature attacks, it gets
// +1/+1 until end of turn for each other attacking <subtype>" as a trigger
// body: the attackers are counted as the ability resolves (CR 611.2c),
// through the text reader so a changed creature type is honoured.
func S82OldB312PlusPerOtherAttacking(name, subtype string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		self, ok := g.LookupCardForEffect(item.SourceCardID)
		if !ok || !SourceIsStillThisPermanent(g, item) || self.AttackingTarget == uuid.Nil {
			return nil
		}
		word := self.TextSubtype(subtype)
		n := 0
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.InstanceID != self.InstanceID && c.AttackingTarget != uuid.Nil && c.HasSubtypeWord(word) {
				n++
			}
		}
		return BoostUntilEOT{Target: self.InstanceID, Power: n, Toughness: n,
			Label: name + " — +1/+1 for each other attacking " + subtype}.Apply(NewContext(g, item))
	}
}
