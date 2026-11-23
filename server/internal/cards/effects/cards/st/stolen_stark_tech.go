package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stolen Stark Tech — {1}{B} Artifact — Equipment (Marvel Super Heroes,
// #4455):
//
//	"Flash
//	 When this Equipment enters, attach it to target creature you control.
//	 That creature gains indestructible until end of turn.
//	 Equipped creature gets +1/+0.
//	 Equip {1}"
//
// The entry trigger is Falcon's Wing Harness' (AttachSourceToTarget) with
// the grant on the same target: the creature gains indestructible even if
// the attach could not happen, because it is a separate instruction.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "51023277-5434-4913-b9dc-df2a3fde1652",
		Name:            "Stolen Stark Tech",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Static:          []game.StaticAbility{effects.PumpAttached(1, 0)},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Stolen Stark Tech — attach it to target creature you control; it gains indestructible until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						if err := effects.AttachSourceToTarget(g, item); err != nil {
							return err
						}
						return effects.FirstTargetGainsUntilEndOfTurn(0, "Stolen Stark Tech — indestructible until end of turn", "indestructible")(g, item)
					}),
				effects.PermanentYouControl("target creature you control", effects.Creature()),
			),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipAbility("{1}"),
		},
	})
}
