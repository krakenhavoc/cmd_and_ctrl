package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vicious Clown — Creature — Human Clown {2}{R}, 2/3 (Duskmourn: House
// of Horror, #4466):
//
//	"Whenever another creature you control with power 2 or less enters,
//	 this creature gets +2/+0 until end of turn."
//
// Snarling Gorehound's trigger condition (the entering creature's power
// as it enters) with Fearless Pup's self pump as the body: the boost
// goes on the Clown itself, and does nothing if it has left (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e02abbc0-687b-4aef-95b8-bf18e199725a",
		Name:         "Vicious Clown",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AnotherCreatureWithPower2OrLessEnteredUnderYourControl,
				"Vicious Clown — this creature gets +2/+0 until end of turn",
				thisGetsUntilEndOfTurn(2, 0, "Vicious Clown — +2/+0")),
		},
	})
}
