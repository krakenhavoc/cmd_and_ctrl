package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shuttle Crew — {2}{W} Creature — Human Officer 1/1 (Star Trek
// Commander, #4974):
//
//	"Flying
//	 When this creature enters, it deals 4 damage to target tapped
//	 creature an opponent controls."
//
// The Wasp, Janet Van Dyne's enters trigger, word for word
// (whenThisEntersFourDamageToTappedOpponentCreature). With no tapped
// creature an opponent controls the trigger is removed (CR 603.3d); one
// that untapped or left by resolution is dealt nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "25855177-a416-4083-a6e4-8c9e00361d41",
		Name:            "Shuttle Crew",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			whenThisEntersFourDamageToTappedOpponentCreature("Shuttle Crew — it deals 4 damage to target tapped creature an opponent controls"),
		},
	})
}
