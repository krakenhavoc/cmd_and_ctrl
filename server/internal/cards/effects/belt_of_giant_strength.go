package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Belt of Giant Strength — Artifact — Equipment {1}{G} (Heroes of the Realm):
//
//	"Equipped creature has base power and toughness 10/10.
//	 Equip {10}. This ability costs {X} less to activate, where X is the power
//	 of the creature it targets."
//
// The base 10/10 is a layer 7b set (SetAttachedBasePT), so counters, anthems
// and other Equipment still add to it. The equip reduction is Warrior's
// Blades' shape: it reads the power of the creature the ability targets as
// the cost is determined (CR 602.2b), through the layers, so a pumped
// creature is cheaper. It never takes the cost below {0} (CR 601.2f).
//
// No simplification.
func init() {
	equip := EquipAbility("{10}")
	equip.CostModifiers = []game.CostModifier{
		CostsLessForTheCardItTargets(
			"This ability costs {X} less to activate, where X is the power of the creature it targets.",
			func(c game.Card) int { return c.CurrentPower() }),
	}
	Register(Spec{
		OracleID:     "eb70939f-c4ce-4224-ab74-acc24771d7f0",
		Name:         "Belt of Giant Strength",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{SetAttachedBasePT(10, 10)},
		Activated:    []ActivatedAbility{equip},
	})
}
