package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Injector Crocodile — Creature — Phyrexian Crocodile {4}{B}{B}, 5/5
// (March of the Machine):
//
//	"When this creature dies, incubate 3.
//	 Swampcycling {2}"
//
// The dies trigger is Infected Defector's; swampcycling is the ordinary
// typecycling activation from hand (cycling.go), "a Swamp card" being
// any card with the Swamp land type.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "619c7d51-51d1-4467-a58b-4f275ff9df0f",
		Name:         "Injector Crocodile",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Injector Crocodile — incubate 3", Do(Incubate{N: 3})),
		},
		Activated: []ActivatedAbility{
			Typecycling("Swampcycling", "{2}", "a Swamp card", IsLandWithSubtype("Swamp")),
		},
	})
}
