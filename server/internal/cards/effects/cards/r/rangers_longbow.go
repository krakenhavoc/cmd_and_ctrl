package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ranger's Longbow — Artifact — Equipment {1}{G} (Adventures in the Forgotten
// Realms, #5499):
//
//	Equipped creature gets +2/+1 and has reach.
//	Equip {3}
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4b9abe7f-050f-4d0a-bc01-f442164f95de",
		Name:         "Ranger's Longbow",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(2, 1), effects.GrantToAttached("reach")},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
