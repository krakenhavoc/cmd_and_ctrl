package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mudbrawler Raiders — Creature — Goblin Warrior {2}{R/G}{R/G}, 3/3
// (Shadowmoor, #5591):
//
//	"This creature can't be blocked by blue creatures."
//
// A block rule, Shield Mare's: a blocker of that colour, as it is now, is
// refused (CR 509.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d8863b64-110d-4f6b-9d0e-4662f2b399cb",
		Name:         "Mudbrawler Raiders",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.And(effects.Creature(), effects.OfColor("U")), "blue creatures"),
		},
	})
}
