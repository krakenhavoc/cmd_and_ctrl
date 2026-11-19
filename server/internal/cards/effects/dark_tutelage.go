package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dark Tutelage — Enchantment {2}{B} (Enchanting Tales):
//
//	"At the beginning of your upkeep, reveal the top card of your library and
//	 put that card into your hand. You lose life equal to its mana value."
//
// Dark Confidant's upkeep flip, word for word, on an enchantment: the reveal
// is not a draw, the mana value is read before the move, a land costs nothing
// and an empty library does nothing (darkConfidantFlip).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0c6ce7c9-316b-4171-a8fa-ca26af1eecb7",
		Name:         "Dark Tutelage",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Dark Tutelage — reveal the top card of your library and put that card into your hand", darkConfidantFlip),
		},
	})
}
