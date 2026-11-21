package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	equip := effects.EquipAbility("{10}")
	equip.CostModifiers = []game.CostModifier{
		effects.CostsLessForTheCardItTargets(
			"This ability costs {X} less to activate, where X is the power of the creature it targets.",
			func(c game.Card) int { return c.CurrentPower() }),
	}
	effects.Register(effects.Spec{
		OracleID:     "eb70939f-c4ce-4224-ab74-acc24771d7f0",
		Name:         "Belt of Giant Strength",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.SetAttachedBasePT(10, 10)},
		Activated:    []effects.ActivatedAbility{equip},
	})
}
