package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Swaggering Corsair — Creature — Human Pirate {2}{R}, 2/2 (Rivals of
// Ixalan, #5531):
//
//	"Raid — This creature enters with a +1/+1 counter on it if you
//	 attacked this turn."
//
// Goblin Boarders' raid: a self-replacement on the entry (CR 614.1c), read
// as the Corsair enters against the turn's attack record.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "48fa47c0-544e-4374-a0a6-ef8f65cdf0dd",
		Name:         "Swaggering Corsair",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.AvB16SelfEntersWithCountersIfRaid(game.CounterPlusOne, 1)},
	})
}
