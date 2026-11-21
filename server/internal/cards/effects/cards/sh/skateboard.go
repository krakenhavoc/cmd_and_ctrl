package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skateboard — {1} Artifact — Equipment (Teenage Mutant Ninja Turtles,
// #4457):
//
//	"When this Equipment enters, tap target permanent.
//	 Equipped creature gets +1/+0 and has haste.
//	 Equip {1}"
//
// Lavaspur Boots's statics and equip (CR 702.6a), plus an enters trigger
// that taps its target (tapTheFirstLegalTarget, Storm Elemental's body).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "decfde5e-ea53-4b15-8f69-51da7ef39d15",
		Name:         "Skateboard",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 0),
			effects.GrantToAttached("haste"),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Skateboard — tap target permanent", effects.TapTheFirstLegalTarget),
				effects.TargetPermanent("target permanent")),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipAbility("{1}"),
		},
	})
}
