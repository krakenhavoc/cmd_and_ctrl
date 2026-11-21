package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ezio, Brash Novice — Legendary Creature — Human {1}{R/W}, 1/1 (Assassin's
// Creed, #5948):
//
//	"Whenever Ezio attacks, put a +1/+1 counter on it.
//	 As long as Ezio has two or more counters on it, it has first strike and
//	 is an Assassin in addition to its other types."
//
// The trigger is a counter on Ezio himself. The second sentence is two statics
// gated on the live count of counters of every kind on him (a layer 6 keyword
// and a layer 4 type addition), read on every recompute (CR 611.3a), so the
// first strike and the type go the moment he is down to one counter. The type
// is added to his others (CR 205.1b), not replacing Human.
//
// No simplification.
func init() {
	twoCounters := func(target *game.Card, _ *game.Game, source *game.Card) bool {
		if target.InstanceID != source.InstanceID {
			return false
		}
		n := 0
		for _, c := range source.Counters {
			n += c
		}
		return n >= 2
	}
	Register(Spec{
		OracleID:     "d4e2b541-31b6-4bcb-a28a-d897deb91484",
		Name:         "Ezio, Brash Novice",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label:     "As long as Ezio has two or more counters on it, it has first strike.",
				Layer:     game.Layer6Ability,
				AppliesTo: twoCounters,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					appendKeywordsTo(c, []string{"first strike"})
				},
			},
			{
				Label:     "As long as Ezio has two or more counters on it, it is an Assassin in addition to its other types.",
				Layer:     game.Layer4Type,
				AppliesTo: twoCounters,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
					if !hasFold(c.Subtypes, source.TextSubtype("Assassin")) {
						c.Subtypes = append(c.Subtypes, source.TextSubtype("Assassin"))
					}
				},
			},
		},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Ezio, Brash Novice — put a +1/+1 counter on it",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
