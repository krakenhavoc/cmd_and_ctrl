package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Worship — Enchantment {3}{W} (Tales of Middle-earth Commander):
//
//	"If you control a creature, damage that would reduce your life total to
//	 less than 1 reduces it to 1 instead."
//
// Sustaining Spirit's life floor, conditioned on its controller controlling
// a creature as the damage is dealt (Spec.PlayerLifeFloorOneIfCreature, read
// off the battlefield in game/life_floor.go). Only damage is floored, the
// damage is still dealt in full (lifelink and commander damage see all of
// it), and a "lose life" effect or a life payment is not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                     "99f9a4c8-b59f-4a41-8577-6c1e4684b240",
		Name:                         "Worship",
		Completeness:                 effects.CompletenessFull,
		PlayerLifeFloorOneIfCreature: true,
	})
}
