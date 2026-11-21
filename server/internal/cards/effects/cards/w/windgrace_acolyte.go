package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Windgrace Acolyte — Creature — Cat Warrior {4}{B}, 3/2 (Dominaria, #5530):
//
//	"Flying
//	 When this creature enters, mill three cards and you gain 3 life."
//
// Shinra Reinforcements' enters trigger with flying: mill three, then
// gain 3 life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "334a98c0-9a16-4145-8f5b-bdf017afe73f",
		Name:            "Windgrace Acolyte",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Windgrace Acolyte — mill three cards and gain 3 life",
				effects.Do(effects.MillCards{N: 3}, effects.GainLife{Amount: 3})),
		},
	})
}
