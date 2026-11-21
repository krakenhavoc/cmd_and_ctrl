package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Winged Boots — Artifact — Equipment {1}{U} (Outlaws of Thunder Junction
// Commander):
//
//	"Equipped creature has flying and ward {4}.
//	 Equip {1}"
//
// Flying is the shared keyword grant; ward is a triggered ability (CR 702.21a)
// the Equipment carries for its current host, the Lavaspur Boots shape.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a178a789-1f5f-46ca-a90b-574bd236a671",
		Name:         "Winged Boots",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.GrantToAttached("flying"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WardAttached(effects.WardMana("{4}"), "Winged Boots — ward {4}"),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipAbility("{1}"),
		},
	})
}
