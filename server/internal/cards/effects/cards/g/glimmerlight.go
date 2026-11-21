package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glimmerlight — Artifact — Equipment {2} (Duskmourn: House of Horror,
// #4466):
//
//	"When this Equipment enters, create a 1/1 white Glimmer enchantment
//	 creature token.
//	 Equipped creature gets +1/+1.
//	 Equip {1} ({1}: Attach to target creature you control. Equip only as
//	 a sorcery.)"
//
// Dwarven Shortsword's shape without the attach: the enters trigger makes
// Grand Entryway's Glimmer, the static is PumpAttached and equip is the
// shared sorcery-speed EquipAbility (CR 702.6a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8fd88b98-9899-4bc0-a4b7-7ef6fba36dfd",
		Name:         "Glimmerlight",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(1, 1)},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Glimmerlight — create a 1/1 white Glimmer enchantment creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Glimmer enchantment"), N: 1})),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{1}")},
	})
}
