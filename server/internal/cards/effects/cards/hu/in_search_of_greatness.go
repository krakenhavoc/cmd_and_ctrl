package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// In Search of Greatness — Enchantment {G}{G}:
//
//	"At the beginning of your upkeep, you may cast a permanent spell from your
//	 hand with mana value equal to 1 plus the greatest mana value among other
//	 permanents you control without paying its mana cost. If you don't, scry 1."
//
// The mana value is counted as the trigger resolves, over the other permanents
// its controller controls (this enchantment not among them), and the offer is the
// free cast of exactly that value among the permanent cards in hand, made during
// the resolution (CR 608.2g, Cecily's offer). A hand with none to cast, or a
// declined offer, scries 1; a cast does not. X in a cost is 0 and extra costs are
// still paid.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a7b45840-505d-4b7d-a084-549be08b0c20",
		Name:         "In Search of Greatness",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 1},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("In Search of Greatness — you may cast a permanent spell from your hand with mana value 1 plus the greatest among your other permanents without paying its mana cost; if you don't, scry 1",
				effects.OldB072Greatness),
		},
	})
}
