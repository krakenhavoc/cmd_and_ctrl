package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vadmir, New Blood — Legendary Creature — Vampire Rogue {1}{B}, 2/2
// (Outlaws of Thunder Junction, #4468):
//
//	"Whenever you commit a crime, put a +1/+1 counter on Vadmir. This
//	 ability triggers only once each turn. (Targeting opponents,
//	 anything they control, and/or cards in their graveyards is a
//	 crime.)
//	 As long as Vadmir has four or more +1/+1 counters on it, it has
//	 menace and lifelink."
//
// The crime trigger is Marauding Sphinx's once-each-turn crime (CR
// 700.13) with a +1/+1 counter on itself as its effect. The keywords are
// a layer 6 static on Vadmir alone, on while it has four or more +1/+1
// counters, whatever put them there. Like Idol of False Gods' statics it
// declares StartsAtCounters, so the effect is timestamped at the fourth
// counter (CR 613.7) and an ability removal that arrived earlier
// applies first.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b000bcf5-529a-447a-bc2e-ae6f42d6073b",
		Name:         "Vadmir, New Blood",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCommitACrimeOnceEachTurn("Vadmir, New Blood — put a +1/+1 counter on Vadmir (crime)",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
		Static: []game.StaticAbility{{
			Label:            "As long as Vadmir has four or more +1/+1 counters on it, it has menace and lifelink.",
			Layer:            game.Layer6Ability,
			AppliesTo:        vadmirHasFourCounters,
			StartsAtCounters: vadmirFourthCounter,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Abilities = game.AppendKeywordAbility(c.Abilities, "menace")
				c.Abilities = game.AppendKeywordAbility(c.Abilities, "lifelink")
			},
		}},
	})
}

// vadmirFourthCounter is the moment the static's effect begins.
var vadmirFourthCounter = game.CounterThreshold{Kind: game.CounterPlusOne, AtLeast: 4}

// vadmirHasFourCounters is the static's condition: Vadmir itself, with
// four or more +1/+1 counters.
func vadmirHasFourCounters(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.InstanceID == source.InstanceID && target.Counters[game.CounterPlusOne] >= 4
}
