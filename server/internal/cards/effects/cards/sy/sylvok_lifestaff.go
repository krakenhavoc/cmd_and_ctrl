package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "e69b8388-1311-427f-9300-42badd5338ff",
		Name:         "Sylvok Lifestaff",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(1, 0)},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.EquippedCreatureDied(ev, source)
			}, "Sylvok Lifestaff — you gain 3 life", effects.Do(effects.GainLife{Amount: 3})),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{1}")},
	})
}
