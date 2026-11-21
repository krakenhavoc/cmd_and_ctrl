package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Panther Habit — Artifact — Equipment {4}:
//
//	"If equipped creature would be dealt damage, prevent that damage and put that many +1/+1 counters on it.
//	 Equip {2}"
//
// ADR 0108 §8 (#1906): a prevention static on the Equipment whose
// additional effect puts "that many" +1/+1 counters on the creature it
// would have been dealt to — the damage, prevented or not, so damage that
// can't be prevented still adds them (CR 615.12). One application per
// recipient in a damage instance, with the instance's total.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "75ec25b3-ae91-46be-af8b-2649727638f4",
		Name:         "Panther Habit",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.PreventDamageDealtTo(effects.PreventionStatic{
				To:    effects.ToEquippedCreature,
				Then:  effects.ThatManyCountersOnItBody,
				Label: "Panther Habit — prevent damage to equipped creature and put that many +1/+1 counters on it",
			}),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
