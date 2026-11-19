package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Star Pupil — {W} Creature — Human Wizard 0/0 (Foundations Jumpstart):
//
//	"This creature enters with a +1/+1 counter on it.
//	 When this creature dies, put its counters on target creature you
//	 control."
//
// Dockworker Drone's shape: the entry counter is a self-replacement, and
// the death trigger reads every counter the source had as it last existed
// (CR 122.8, 603.10a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "37fd3a9d-74fe-4180-805a-b865e7228a77",
		Name:         "Star Pupil",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterPlusOne, 1, "Star Pupil: enters with a +1/+1 counter on it"),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisDies("Star Pupil — put its counters on target creature you control",
				PutItsCountersOnTargetCreature),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}
