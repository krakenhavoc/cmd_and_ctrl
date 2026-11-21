package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Falcon's Wing Harness — {1}{U} Artifact — Equipment:
//
//	"When this Equipment enters, attach it to target creature you control.
//	 Equipped creature gets +1/+1 and has flying and ward {1}.
//	 Equip {2}{U}"
//
// The entry trigger is Maul of the Skyclaves' shape; the ward is granted to
// the equipped creature the way Armguard Familiar grants its own.
//
// No simplification. Marvel Super Heroes (#4455).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e69db5f0-beed-48e8-bc25-768b1e8be5ea",
		Name:         "Falcon's Wing Harness",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.GrantToAttached("flying"),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Falcon's Wing Harness — attach it to target creature you control", effects.AttachSourceToTarget),
				effects.PermanentYouControl("target creature you control", effects.Creature()),
			),
			effects.WardAttached(effects.WardMana("{1}"), "Falcon's Wing Harness — equipped creature's ward {1}"),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipAbility("{2}{U}"),
		},
	})
}
