package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sylvok Lifestaff — Artifact — Equipment {1} (Scars of Mirrodin, #5577):
//
//	"Equipped creature gets +1/+0.
//	 Whenever equipped creature dies, you gain 3 life.
//	 Equip {1}"
//
// Lead Pipe's shape: the pump, a leaves-the-battlefield trigger read
// off the equipped creature as it last existed (equippedCreatureDied),
// and the equip row. The Lifestaff's controller gains the life.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e69b8388-1311-427f-9300-42badd5338ff",
		Name:         "Sylvok Lifestaff",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(1, 0)},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return equippedCreatureDied(ev, source)
			}, "Sylvok Lifestaff — you gain 3 life", Do(GainLife{Amount: 3})),
		},
		Activated: []ActivatedAbility{EquipAbility("{1}")},
	})
}
