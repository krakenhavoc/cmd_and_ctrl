package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Candlegrove Witch — Creature — Human Warlock {1}{W}, 2/2 (Midnight
// Hunt, #5497):
//
//	"Coven — At the beginning of combat on your turn, if you control
//	 three or more creatures with different powers, this creature gains
//	 flying until end of turn."
//
// The intervening if (CR 603.4) is trekB48Coven, checked as combat begins
// and as the trigger resolves; the grant is a layer 6 effect pinned to this
// object (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "95f47e9b-0290-4ec4-a1ba-bc5a432ffab2",
		Name:         "Candlegrove Witch",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.MidB05CovenThisGainsKeyword("Candlegrove Witch", "flying")},
	})
}
