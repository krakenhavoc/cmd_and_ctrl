package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Murderer's Axe — Artifact — Equipment {4} (Shadows over Innistrad, #5543):
//
//	"Equipped creature gets +2/+2.
//	 Equip—Discard a card."
//
// Equip is an activated ability (CR 702.6a) whose whole cost is the discard,
// paid as the ability is activated and refused with an empty hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f0ea953a-10c3-4ec7-8f03-f7aed75dbb94",
		Name:         "Murderer's Axe",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(2, 2)},
		Activated:    []effects.ActivatedAbility{effects.EquipPayingAbility("Equip—Discard a card", effects.DiscardACard())},
	})
}
