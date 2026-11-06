package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "6cfe63ac-3581-461b-8ed3-3e231851e3f8",
		Name:         "Shinra Reinforcements",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Shinra Reinforcements — mill three cards and gain 3 life",
				Do(MillCards{N: 3}, GainLife{Amount: 3})),
		},
	})
}
