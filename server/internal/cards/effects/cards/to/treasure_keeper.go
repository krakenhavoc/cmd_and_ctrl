package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Treasure Keeper — Artifact Creature — Construct {4}, 3/3 (Foundations
// Jumpstart):
//
//	"When this creature dies, reveal cards from the top of your library until
//	 you reveal a nonland card with mana value 3 or less. You may cast that
//	 card without paying its mana cost. Put all revealed cards not cast this
//	 way on the bottom of your library in a random order."
//
// The reveal is revealUntil; the free cast is offered from the library while
// the trigger resolves (CR 608.2g, resolution_cast prompt), and every
// revealed card not cast goes to the bottom in a random order. A library with
// no such card reveals itself whole and bottoms all of it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3ef7e4c3-f87b-442d-8680-39673ede07e3",
		Name:         "Treasure Keeper",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Treasure Keeper — reveal until a nonland card with mana value 3 or less, you may cast it free, the rest on the bottom",
				effects.TreasureKeeperReveal),
		},
	})
}
