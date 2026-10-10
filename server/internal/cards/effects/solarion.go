package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Solarion — Artifact Creature — Construct {7}, 0/0:
//
//	"Sunburst (This creature enters with a +1/+1 counter on it for each
//	 color of mana spent to cast it.)
//	 {T}: Double the number of +1/+1 counters on this creature."
//
// Sunburst is a keyword the engine reads off the resolving spell (ADR
// 0109 §11, #1552). "Double" is putting as many +1/+1 counters as it
// has (CR 701.10e), read as the ability resolves, so a counter doubler
// applies to the ones put on.
func init() {
	Register(Spec{
		OracleID:            "4022e540-8fa0-4eee-bcf1-df5046289f2b",
		Name:                "Solarion",
		Completeness:        CompletenessCaveats,
		Caveats:             []string{"With strict mana off, the game doesn't track which mana you spent, so it enters with no counters at all."},
		WantsDistinctColors: true,
		PrintedKeywords:     []string{game.KeywordSunburst},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Double the number of +1/+1 counters on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return b23DoublePlusOneCountersOn(NewContext(g, item), item.SourceCardID)
			},
		}},
	})
}
