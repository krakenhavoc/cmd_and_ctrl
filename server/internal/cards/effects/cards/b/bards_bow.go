package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bard's Bow — Artifact — Equipment {2}{G} (Final Fantasy, #4462):
//
//	"Job select (When this Equipment enters, create a 1/1 colorless Hero
//	 creature token, then attach this to it.)
//	 Equipped creature gets +2/+2, has reach, and is a Bard in addition
//	 to its other types.
//	 Perseus's Bow — Equip {6}"
//
// The job-select Equipment shape (Sage's Nouliths, Dancer's Chakrams):
// the shared JobSelect trigger, a layer 7c pump, a layer 6 keyword grant
// and a layer 4 creature-type addition. "Perseus's Bow" is a flavor word
// with no rules meaning; the ability is an ordinary equip.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4aceb786-960d-4719-9de8-79dffd491a51",
		Name:         "Bard's Bow",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GrantToAttached("reach"),
			effects.AttachedIsAlso("Bard"),
		},
		Triggered: []game.TriggeredAbility{effects.JobSelect("Bard's Bow")},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{6}")},
	})
}
