package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "48fa47c0-544e-4374-a0a6-ef8f65cdf0dd",
		Name:         "Swaggering Corsair",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{avB16SelfEntersWithCountersIfRaid(game.CounterPlusOne, 1)},
	})
}
