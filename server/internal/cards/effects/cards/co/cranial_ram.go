package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cranial Ram — Artifact — Equipment {B}{R} (Modern Horizons 3, #5948):
//
//	"Living weapon (When this Equipment enters, create a 0/0 black Phyrexian
//	 Germ creature token, then attach this to it.)
//	 Equipped creature gets +X/+1, where X is the number of artifacts you
//	 control.
//	 Equip {2}"
//
// The shared LivingWeapon trigger, Cranial Plating's count (every artifact
// its controller controls, the Ram itself and an animated permanent included,
// read after the earlier layers) at layer 7c, and a flat +0/+1 beside it. The
// Germ is a 0/0 that the Ram makes a 1/1 plus one for each artifact, so it
// survives while the Ram is attached.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e958b3c5-0666-4a99-a933-e5cb8b0da9da",
		Name:         "Cranial Ram",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.LivingWeapon("Cranial Ram"),
		},
		Static: []game.StaticAbility{
			effects.PumpAttachedPer(1, 0, effects.ArtifactsControlledBy),
			effects.PumpAttached(0, 1),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipAbility("{2}"),
		},
	})
}
