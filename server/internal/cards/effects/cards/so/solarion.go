package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:            "4022e540-8fa0-4eee-bcf1-df5046289f2b",
		Name:                "Solarion",
		ReadsManaSpent:      game.ReadsOwnPayment,
		Completeness:        effects.CompletenessFull,
		WantsDistinctColors: true,
		PrintedKeywords:     []string{game.KeywordSunburst},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Double the number of +1/+1 counters on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.B23DoublePlusOneCountersOn(effects.NewContext(g, item), item.SourceCardID)
			},
		}},
	})
}
