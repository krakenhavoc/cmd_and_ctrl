package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Centaur Peacemaker — Creature — Centaur Cleric {1}{G}{W}, 3/3 (Guilds of Ravnica, #5525):
//
//	"When this creature enters, each player gains 4 life."
//
// The enter trigger is Putrid Warrior's "each player gains" body
// (eachPlayerLifeChanges): every player still in the game gains 4.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "495d95cf-dcce-4a15-a544-79aced7ef98d",
		Name:         "Centaur Peacemaker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Centaur Peacemaker — each player gains 4 life",
				Do(Instruction(func(ctx *Context) error { return eachPlayerLifeChanges(ctx, 4) }))),
		},
	})
}
