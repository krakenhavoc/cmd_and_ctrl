package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blood-Toll Harpy — Creature — Harpy {2}{B}, 2/1 (Theros, #5561):
//
//	"Flying
//	 When this creature enters, each player loses 1 life."
//
// The enters trigger takes 1 life from every player at the table, its
// controller included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "98f9619e-585f-486f-b143-fb42085eb583",
		Name:            "Blood-Toll Harpy",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Blood-Toll Harpy — each player loses 1 life", effects.Do(effects.Instruction(effects.SocB22EachPlayerLosesOneLife))),
		},
	})
}
