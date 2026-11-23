package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shinra Reinforcements — Creature — Human Soldier {2}{B}, 2/3 (Final
// Fantasy, #4462):
//
//	"When this creature enters, mill three cards and you gain 3 life.
//	 (To mill three cards, put the top three cards of your library into
//	 your graveyard.)"
//
// One enters trigger, two steps in printed order: the controller mills
// three (fewer from a shorter library) and then gains 3 life, which does
// not depend on the mill.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6cfe63ac-3581-461b-8ed3-3e231851e3f8",
		Name:         "Shinra Reinforcements",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Shinra Reinforcements — mill three cards and gain 3 life",
				effects.Do(effects.MillCards{N: 3}, effects.GainLife{Amount: 3})),
		},
	})
}
