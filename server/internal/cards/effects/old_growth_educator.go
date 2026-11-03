package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Old-Growth Educator — Creature — Treefolk Druid {2}{B}{G}, 4/4 (Secrets of
// Strixhaven, #4456):
//
//	"Reach, vigilance
//	 Infusion — When this creature enters, put two +1/+1 counters on it if
//	 you gained life this turn."
//
// The "if" is not an intervening one: the ability triggers on every entry
// and checks the life gained as it resolves (the printed ruling), so a life
// gain in response to the trigger counts. The counters go on this creature
// only while it is still the permanent that entered.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2b7362a9-3518-4071-ab1b-40689a50e12a",
		Name:            "Old-Growth Educator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "vigilance"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Old-Growth Educator — Infusion: two +1/+1 counters on it if you gained life this turn",
				Do(If{
					Cond: CondLifeGained{},
					Then: []Applier{CounterOnThis{Kind: game.CounterPlusOne, N: 2}},
				})),
		},
	})
}
