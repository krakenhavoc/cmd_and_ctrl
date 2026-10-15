package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bureau Headmaster — Creature — Human Assassin {R}{W}, 2/2 (#3778):
//
//	"Equipment spells you cast cost {1} less to cast.
//	 Equip abilities you activate cost {1} less to activate."
//
// Two modifiers: the cast one prices spells, the activation one prices
// equip abilities (the equip bit on the activation's cost query), and
// neither reaches the other's announcements.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c022b5df-1d85-44c0-94a1-d01b345b845b",
		Name:         "Bureau Headmaster",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Equipment spells you cast cost {1} less to cast.",
				YourSpell(), equipmentSpell()),
			EquipAbilitiesYouActivateCostLess(1, "Equip abilities you activate cost {1} less to activate."),
		},
	})
}

func equipmentSpell() CostPredicate {
	return func(q game.CostQuery) bool { return q.Card.HasSubtype("Equipment") }
}
