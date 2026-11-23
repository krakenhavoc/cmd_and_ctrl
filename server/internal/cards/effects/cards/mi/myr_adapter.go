package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Myr Adapter — Artifact Creature — Myr {3}, 1/1 (Mirrodin, #5606):
//
//	"This creature gets +1/+1 for each Equipment attached to it."
//
// Loxodon Punisher's count at +1/+1 (mrdB5PerEquipmentOnThis): the
// Equipment attached to the Adapter itself, read live.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "714f916a-10ac-4c54-b1c7-a97b7ad8b9e5",
		Name:         "Myr Adapter",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.MrdB5PerEquipmentOnThis(1)},
	})
}
