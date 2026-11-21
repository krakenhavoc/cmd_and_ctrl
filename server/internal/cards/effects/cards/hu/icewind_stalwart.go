package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Icewind Stalwart — Creature — Tiefling Warrior {3}{W}, 3/3 (Commander
// Legends: Battle for Baldur's Gate, #5948):
//
//	"Protection Fighting Style — When this creature enters, exile up to one
//	 target non-Warrior creature you control, then return it to the
//	 battlefield under its owner's control."
//
// The Mind Stone's blink with an "up to one" target: the card comes back as
// a new object under its owner's control (CR 400.7), so it re-triggers every
// "enters" ability it has. "Non-Warrior" is read through the layered
// subtypes. A target gone by resolution is a no-op (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9ad12c75-e97d-404a-bfa9-26ff1d0bb506",
		Name:         "Icewind Stalwart",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Icewind Stalwart — Protection Fighting Style: exile up to one target non-Warrior creature you control, then return it to the battlefield under its owner's control",
					effects.FlickerFirstLegalTargetUnderOwner),
				effects.TargetCreature("up to one target non-Warrior creature you control", effects.YouControl(), effects.Not(effects.Subtype("Warrior"))).WithCount(0, 1)),
		},
	})
}
