package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Knowledge Is Power — Enchantment {3}{W}{U} (Murders at Karlov Manor
// Commander, #5368):
//
//	"Creatures you control get +X/+X, where X is the number of cards
//	 you've drawn this turn."
//
// A layer 7c anthem over the turn's draw tally, which counts every card
// drawn, including one since discarded or played. The static declares
// that a draw drops the layer cache (a draw moves nothing on the
// battlefield), so the bonus is current the moment the card is drawn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8806ee05-10d5-4492-b672-2f6bcab94be0",
		Name:         "Knowledge Is Power",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{cmdr2S22PlusXForCardsDrawnThisTurn()},
	})
}
