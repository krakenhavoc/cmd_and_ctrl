package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kor Celebrant — Creature — Kor Cleric {2}{W}, 1/4 (Zendikar Rising,
// #5508):
//
//	"Whenever this creature or another creature you control enters, you
//	 gain 1 life."
//
// Bogwater Lumaret's trigger: one per creature entering under your
// control, the Celebrant itself included, read off the entering object's
// layered type, so an opponent's creature does not count.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3f8f0ce9-d2ce-45cc-9fbf-0e2819b07a3e",
		Name:         "Kor Celebrant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, CreatureEnteredUnderYourControl,
				"Kor Celebrant — you gain 1 life", Do(GainLife{Amount: 1})),
		},
	})
}
