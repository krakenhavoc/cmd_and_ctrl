package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tricorder — Artifact — Equipment {1} (Star Trek, #4973):
//
//	"Equipped creature gets +1/+1.
//	 Whenever equipped creature attacks, surveil 1.
//	 Equip {1}"
//
// PumpAttached(1, 1), EquipAbility("{1}") and Argentum Armor's
// equipped-creature attack trigger (attachedCreatureAttacked) with a
// surveil 1 for the Equipment's controller. An attack at a planeswalker
// or a battle is still an attack, so it triggers then too.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1385679f-bdcd-4d7e-8266-1df0e36cc2da",
		Name:         "Tricorder",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(1, 1)},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attachedCreatureAttacked(ev, source)
			}, "Tricorder — equipped creature attacked: surveil 1", Do(Surveil{N: 1})),
		},
		Activated: []ActivatedAbility{EquipAbility("{1}")},
	})
}
