package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Barrier of Bones — Creature — Skeleton Wall {B}, 0/3 (Guilds of Ravnica, #5525):
//
//	"Defender
//	 When this creature enters, surveil 1."
//
// Defender is printed data; the enter trigger is Coin of Fate's surveil 1
// on the stack (CR 701.25a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4f25b125-1ded-4a34-9ed1-e6ce087ec48d",
		Name:            "Barrier of Bones",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Barrier of Bones — surveil 1", Do(Surveil{N: 1})),
		},
	})
}
