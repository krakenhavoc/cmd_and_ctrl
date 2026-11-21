package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Monk's Fist — Artifact — Equipment {2} (Final Fantasy, #4462):
//
//	"Job select (When this Equipment enters, create a 1/1 colorless Hero
//	 creature token, then attach this to it.)
//	 Equipped creature gets +1/+0 and is a Monk in addition to its other
//	 types.
//	 Equip {2}"
//
// Bard's Bow's job-select shape with no keyword: the shared JobSelect
// trigger, a layer 7c pump and a layer 4 creature-type addition.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1b524149-340e-4f94-9520-6a5258fabb00",
		Name:         "Monk's Fist",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 0),
			effects.AttachedIsAlso("Monk"),
		},
		Triggered: []game.TriggeredAbility{effects.JobSelect("Monk's Fist")},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
