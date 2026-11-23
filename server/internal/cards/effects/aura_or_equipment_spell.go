package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// AuraOrEquipmentSpell is the cost predicate "Aura and Equipment spells":
// the spell has the Aura or the Equipment subtype. Danitha Capashen,
// Paragon and Golden-Tail Trainer share it.
func AuraOrEquipmentSpell() CostPredicate {
	return func(q game.CostQuery) bool {
		return q.Card.HasSubtype("Aura") || q.Card.HasSubtype("Equipment")
	}
}
