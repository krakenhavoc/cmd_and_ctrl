package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ascendant Packleader — Creature — Wolf {G}, 2/1 (Innistrad: Crimson Vow,
// #5495):
//
//	"This creature enters with a +1/+1 counter on it if you control a
//	 permanent with mana value 4 or greater.
//	 Whenever you cast a spell with mana value 4 or greater, put a +1/+1
//	 counter on this creature."
//
// The entry is a CR 614.1c self-replacement counted as the Packleader
// enters (Cindering Cutthroat's shape): the permanents are read through
// the layers and the Packleader is never its own permanent. The cast
// trigger reads the spell's mana value on the stack, so X counts at its
// announced value (Enraged Flamecaster's trigger).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7e81853c-d0c0-40bc-80fc-13a5964878b2",
		Name:         "Ascendant Packleader",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b19EntersWithCountersCounted(game.CounterPlusOne, vowB07OneIfYouControlManaValueFourOrMore,
				"Ascendant Packleader enters with a +1/+1 counter on it if you control a permanent with mana value 4 or greater"),
		},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(ManaValueGE(4), "Ascendant Packleader — put a +1/+1 counter on this creature",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
