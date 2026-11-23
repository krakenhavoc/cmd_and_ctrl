package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Coral Sword — Artifact — Equipment {R} (Final Fantasy, #4462):
//
//	"Flash
//	 When this Equipment enters, attach it to target creature you
//	 control. That creature gains first strike until end of turn.
//	 Equipped creature gets +1/+0.
//	 Equip {1}"
//
// Barbed Bloodletter's shape with first strike for the rider
// (fin_b3AttachAndGrantUntilEOT). "That creature" is the target, not the
// equipped creature: if the Sword is gone in response the attach does
// nothing (CR 701.3b) and the still-legal creature still gains first
// strike.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d4f0dd14-bc2a-4c6d-bfd5-bc5838ab18a5",
		Name:            "Coral Sword",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 0),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Coral Sword — attach it to target creature you control; it gains first strike until end of turn",
					effects.FinB3AttachAndGrantUntilEOT("Coral Sword", "first strike")),
				effects.PermanentYouControl("target creature you control", effects.Creature()),
			),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipAbility("{1}"),
		},
	})
}
