package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chaos Moon — Enchantment {3}{R} (Ice Age, S82 OLD, #5636):
//
//	"At the beginning of each upkeep, count the number of permanents. If the
//	 number is odd, until end of turn, red creatures get +1/+1 and whenever a
//	 player taps a Mountain for mana, that player adds an additional {R}. If the
//	 number is even, until end of turn, red creatures get -1/-1 and if a player
//	 taps a Mountain for mana, that Mountain produces colorless mana instead of
//	 any other type."
//
// Counted as the trigger resolves. The odd side is a pump of the red creatures
// then on the battlefield (CR 611.2c) and Bubbling Muck's turn-long mana trigger
// for Mountains; the even side is the -1/-1 and a CR 614.1a replacement on mana
// production for every player's Mountains (game.ModLandSubtypeManaColorless),
// Pale Moon's shape, keeping the amount.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "728e2960-1d85-4b3f-8b5b-915f9b9ffd2f",
		Name:         "Chaos Moon",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtEachUpkeep("Chaos Moon — count the permanents: odd, red creatures get +1/+1 and Mountains add {R}; even, red creatures get -1/-1 and Mountains make colorless mana",
				effects.OldB834ChaosMoon),
		},
	})
}
