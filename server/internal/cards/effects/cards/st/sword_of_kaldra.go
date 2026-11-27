package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sword of Kaldra — Legendary Artifact — Equipment {4} (Mirrodin, #5606):
//
//	"Equipped creature gets +5/+5.
//	 Whenever equipped creature deals damage to a creature, exile that
//	 creature. (Exile it only if it's still on the battlefield.)
//	 Equip {4}"
//
// Kaldra Compleat's exile, on the Equipment itself and for damage of any kind
// (CR 120.2): the victim is named as the object it was when the damage was
// dealt, so one that died, or died and came back, is not exiled (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "35f63c60-b0a1-459d-bf51-a8549fb06930",
		Name:         "Sword of Kaldra",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(5, 5)},
		Triggered: []game.TriggeredAbility{
			effects.OldB605WheneverEquippedCreatureDealsDamageToACreature("Sword of Kaldra — exile the creature equipped creature dealt damage to"),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{4}")},
	})
}
