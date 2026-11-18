package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brazen Dwarf — Creature — Dwarf Shaman {1}{R}, 1/3 (Adventures in the
// Forgotten Realms):
//
//	"Whenever you roll one or more dice, this creature deals 1 damage to
//	 each opponent."
//
// One trigger per instruction to roll, however many dice (WheneverYouRollDice).
func init() {
	Register(Spec{
		OracleID: "256ab1e2-0c6a-41c1-ba01-021f3c90cfa2", Name: "Brazen Dwarf", Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			afrWheneverYouRollDiceDo("Brazen Dwarf — 1 damage to each opponent", Do(DamageEachOpponent{N: 1})),
		},
	})
}
