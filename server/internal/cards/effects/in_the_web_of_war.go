package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// In the Web of War — Enchantment {3}{R}{R} (Betrayers of Kamigawa,
// #5602):
//
//	"Whenever a creature you control enters, it gets +2/+0 and gains
//	 haste until end of turn."
//
// Ogre Battledriver's trigger without "another" (the enchantment is never
// a creature itself): the creature that entered is the trigger's object,
// and the boost and haste go to it only while it is still on the
// battlefield as the ability resolves (enteredCreatureGetsTwoAndHaste).
// Tokens count.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9629e6b7-75e1-4447-bf20-4d4a43d6f60a",
		Name:         "In the Web of War",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, CreatureEnteredUnderYourControl,
				"In the Web of War — it gets +2/+0 and gains haste until end of turn",
				enteredCreatureGetsTwoAndHaste("In the Web of War")),
		},
	})
}
