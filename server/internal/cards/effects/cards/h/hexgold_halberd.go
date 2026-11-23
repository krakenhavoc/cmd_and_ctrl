package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hexgold Halberd — Artifact — Equipment {1}{R} (Phyrexia: All Will Be One):
//
//	"For Mirrodin! (When this Equipment enters, create a 2/2 red Rebel
//	 creature token, then attach this to it.)
//	 During your turn, equipped creature has first strike and trample.
//	 Equip {2}{R}"
//
// #5654, S82. The shared ForMirrodin trigger and one equip ability. The
// grant is Quick-Draw Katana's: "your turn" is the Equipment's
// controller's turn (CR 109.5), read on every layer pass, so the keywords
// come and go with the turn and not with the host's controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c43b176a-9001-4fe4-bd9b-809880134067",
		Name:         "Hexgold Halberd",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.ForMirrodin("Hexgold Halberd")},
		Static: []game.StaticAbility{
			effects.GrantToAttachedWhile(effects.DuringSourceControllersTurn, "first strike", "trample"),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}{R}")},
	})
}
