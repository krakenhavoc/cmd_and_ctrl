package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Toggo, Goblin Weaponsmith — Legendary Creature — Goblin Artificer {2}{R},
// 2/2 (Commander Legends):
//
//	"Landfall — Whenever a land you control enters, create a colorless
//	 Equipment artifact token named Rock with "Equipped creature has '{1},
//	 {T}, Sacrifice Rock: This creature deals 2 damage to any target'" and
//	 equip {1}.
//	 Partner"
//
// Landfall is the usual land-entered trigger. The Rock is a token template
// of its own (s82_old_b077_helpers.go) whose granted ability belongs to the
// equipped creature, so the creature taps to pay and deals the damage
// itself. Partner is a deck-construction rule (CR 702.124).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d334717a-f3a8-4ffa-bdc9-44ae7bdcc772",
		Name:         "Toggo, Goblin Weaponsmith",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Landfall("Toggo, Goblin Weaponsmith — create a Rock Equipment token",
				effects.Do(effects.CreateToken{Template: effects.RockEquipmentToken(), N: 1})),
		},
	})
}
