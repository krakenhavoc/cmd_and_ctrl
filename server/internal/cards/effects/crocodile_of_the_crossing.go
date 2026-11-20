package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crocodile of the Crossing — Creature — Crocodile {3}{G}, 5/4
// (Amonkhet, #5535):
//
//	"Haste
//	 When this creature enters, put a -1/-1 counter on target creature you control."
//
// Channeler Initiate's enters trigger with one counter; the target may
// be this creature itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e5e31d27-3907-4c65-a988-a754f217e061",
		Name:            "Crocodile of the Crossing",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Crocodile of the Crossing — put a -1/-1 counter on target creature you control", cmdr2B25CountersOnTarget(1)),
				TargetCreature("target creature you control", YouControl()),
			),
		},
	})
}
