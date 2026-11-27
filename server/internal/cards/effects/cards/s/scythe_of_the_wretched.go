package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scythe of the Wretched — Artifact — Equipment {2} (Mirrodin, #5606):
//
//	"Equipped creature gets +2/+2.
//	 Whenever a creature dealt damage by equipped creature this turn dies,
//	 return that card to the battlefield under your control. Attach this
//	 Equipment to that creature.
//	 Equip {4}"
//
// Zurgo Helmsmasher's "dealt damage by this creature this turn" record
// (CR 400.7: the damage counts for the object that dealt it), read for the
// creature this Equipment is attached to when the creature dies. The card
// comes back under the Equipment controller's control, as a new object, and the Equipment is attached to it. A card that
// has left the graveyard by then is not returned, and an Equipment that has
// left the battlefield attaches nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9dc26b9f-3b02-46f2-88f8-eb6dffa04ee5",
		Name:         "Scythe of the Wretched",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(2, 2)},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, effects.OldB605DealtDamageByEquippedCreatureDied,
				"Scythe of the Wretched — return that card to the battlefield under your control and attach this Equipment to it", effects.OldB605ReturnTheDeadAndAttach),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{4}")},
	})
}
