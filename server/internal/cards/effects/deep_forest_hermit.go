package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deep Forest Hermit — Creature — Elf Druid {3}{G}{G}, 1/1:
//
//	"Vanishing 3 (This creature enters with three time counters on it.
//	 At the beginning of your upkeep, remove a time counter from it.
//	 When the last is removed, sacrifice it.)
//	 When this creature enters, create four 1/1 green Squirrel creature
//	 tokens.
//	 Squirrels you control get +1/+1."
//
// Vanishing is cmdr2B17Vanishing*: the entry counters are a
// replacement, the upkeep removal and the last-counter sacrifice are
// two triggers (CR 702.63a). The Hermit is itself not a Squirrel.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "80e9a42e-f4f4-4bfc-91a2-ad124737dca2",
		Name:         "Deep Forest Hermit",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 4, CreatureTokens: 4},
		Replacements: []game.ReplacementEffect{cmdr2B17VanishingEntry(3)},
		Triggered: append(cmdr2B17VanishingTriggers(),
			WhenThisEnters("Deep Forest Hermit — create four 1/1 green Squirrel creature tokens",
				Do(CreateToken{Template: TokenCard("1/1 green Squirrel"), N: 4}))),
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Squirrel"}, YoursOnly: true}, 1, 1),
		},
	})
}
