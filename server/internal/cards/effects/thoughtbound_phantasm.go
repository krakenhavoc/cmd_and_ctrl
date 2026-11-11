package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thoughtbound Phantasm — Creature — Spirit {U}, 2/2:
//
//	"Defender
//	 Whenever you surveil, put a +1/+1 counter on this creature.
//	 As long as this creature has three or more +1/+1 counters on it, it
//	 can attack as though it didn't have defender."
//
// Dimir Spybug's counter trigger. The static half is a defender waiver
// (avB18CanAttackAsThoughNoDefenderWhile, Demon Wall's shape) conditioned on the +1/+1 counters it carries, read
// as the attack is declared.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ca14f1b6-e8c2-4ea5-9375-5ed63f355052",
		Name:            "Thoughtbound Phantasm",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Static: []game.StaticAbility{avB18CanAttackAsThoughNoDefenderWhile(
			"As long as this creature has three or more +1/+1 counters on it, it can attack as though it didn't have defender.",
			func(_ *game.Game, source *game.Card) bool {
				return source != nil && source.Counters[game.CounterPlusOne] >= 3
			})},
		Triggered: []game.TriggeredAbility{
			cmdr2B27WheneverYouSurveil("Thoughtbound Phantasm — put a +1/+1 counter on this creature",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
