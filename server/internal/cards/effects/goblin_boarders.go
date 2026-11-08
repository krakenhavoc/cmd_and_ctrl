package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goblin Boarders — Creature — Goblin Pirate {2}{R}, 3/2 (Foundations,
// #4465):
//
//	"Raid — This creature enters with a +1/+1 counter on it if you
//	 attacked this turn."
//
// Cruel Administrator's raid: a self-replacement on the entry (CR 614.1c),
// read as the Goblin enters. "You attacked this turn" is the turn's attack
// tally, so any attack you declared this turn counts, whatever has happened
// to the attackers since.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4c64feae-95d7-426d-9e42-d3b0841d2054",
		Name:         "Goblin Boarders",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{avB16SelfEntersWithCountersIfRaid(game.CounterPlusOne, 1)},
	})
}
