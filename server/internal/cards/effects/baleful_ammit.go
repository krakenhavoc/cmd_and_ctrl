package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Baleful Ammit — Creature — Crocodile Demon {2}{B}, 4/3
// (Amonkhet, #5535):
//
//	"Lifelink
//	 When this creature enters, put a -1/-1 counter on target creature you control."
//
// Channeler Initiate's enters trigger with one counter; the target may
// be this creature itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1cfa9f7d-ef1f-4907-aeb1-15c0df33fccc",
		Name:            "Baleful Ammit",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Baleful Ammit — put a -1/-1 counter on target creature you control", cmdr2B25CountersOnTarget(1)),
				TargetCreature("target creature you control", YouControl()),
			),
		},
	})
}
