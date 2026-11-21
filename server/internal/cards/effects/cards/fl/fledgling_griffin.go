package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fledgling Griffin — Creature — Griffin {1}{W}, 2/2 (Worldwake, #5582):
//
//	"Landfall — Whenever a land you control enters, this creature gains
//	 flying until end of turn."
//
// Landfall with a layer-6 grant pinned to the creature until the
// turn ends (thisGainsKeywordUntilEndOfTurn).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "314748a6-4e7f-4afb-b77e-5ef366e02b90",
		Name:         "Fledgling Griffin",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Landfall("Fledgling Griffin — gains flying until end of turn (landfall)",
				effects.ThisGainsKeywordUntilEndOfTurn("flying", "Fledgling Griffin — flying until end of turn")),
		},
	})
}
