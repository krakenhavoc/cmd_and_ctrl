package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "334a98c0-9a16-4145-8f5b-bdf017afe73f",
		Name:            "Windgrace Acolyte",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Windgrace Acolyte — mill three cards and gain 3 life",
				Do(MillCards{N: 3}, GainLife{Amount: 3})),
		},
	})
}
