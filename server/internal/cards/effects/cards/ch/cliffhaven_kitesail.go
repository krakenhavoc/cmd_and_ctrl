package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cliffhaven Kitesail — Artifact — Equipment {1} (Foundations Jumpstart):
//
//	"When this Equipment enters, attach it to target creature you control.
//	 Equipped creature has flying.
//	 Equip {2}"
//
// Barbed Bloodletter's entry attach with a flying grant.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "68fb8f21-4403-4829-8826-f2418664197e",
		Name:         "Cliffhaven Kitesail",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.GrantToAttached("flying")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Cliffhaven Kitesail — attach it to target creature you control", effects.AttachSourceToTarget),
				effects.PermanentYouControl("target creature you control", effects.Creature()),
			),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
