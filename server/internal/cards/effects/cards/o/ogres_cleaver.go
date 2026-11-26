package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ogre's Cleaver — Artifact — Equipment {2} (Rise of the Eldrazi, S82 OLD b486):
//
//	"Equipped creature gets +5/+0.
//	 Equip {5}"
//
// A layer-7c pump that follows the attachment (PumpAttached, CR 613.4c) and
// the shared Equip row (sorcery speed, your creature, CR 702.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e32c2ed2-be8b-497c-8b94-9c104f203b49",
		Name:         "Ogre's Cleaver",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(5, 0)},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{5}")},
	})
}
