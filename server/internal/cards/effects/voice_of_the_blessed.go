package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Voice of the Blessed — Creature — Spirit Cleric {W}{W}, 2/2 (Innistrad:
// Crimson Vow, #5495):
//
//	"Whenever you gain life, put a +1/+1 counter on this creature.
//	 As long as this creature has four or more +1/+1 counters on it, it
//	 has flying and vigilance.
//	 As long as this creature has ten or more +1/+1 counters on it, it has
//	 indestructible."
//
// The trigger is Ajani's Pridemate's (one counter per life-gain event, not
// per point). The three keywords are self-only layer-6 grants gated on the
// +1/+1 counters it has right now (Tom Bombadil's KeywordGrant shape), read
// live, so they stop the moment a counter is removed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "01e742ed-660c-43c6-b393-e589a54cbe1b",
		Name:         "Voice of the Blessed",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			KeywordGrant(voiceOfTheBlessedCounters(4), "flying"),
			KeywordGrant(voiceOfTheBlessedCounters(4), "vigilance"),
			KeywordGrant(voiceOfTheBlessedCounters(10), "indestructible"),
		},
		Triggered: []game.TriggeredAbility{
			LifeGainPaying(
				WheneverYouGainLife("Voice of the Blessed — put a +1/+1 counter on it", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
				game.LifeGainPayoff{Counters: 1}),
		},
	})
}

// voiceOfTheBlessedCounters is "as long as this creature has n or more
// +1/+1 counters on it".
func voiceOfTheBlessedCounters(n int) func(target *game.Card, g *game.Game, source *game.Card) bool {
	return func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID && target.Counters[game.CounterPlusOne] >= n
	}
}
