package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sawtooth Thresher — Artifact Creature — Construct {6}, 1/1:
//
//	"Sunburst (This creature enters with a +1/+1 counter on it for each
//	 color of mana spent to cast it.)
//	 Remove two +1/+1 counters from this creature: It gets +4/+4 until
//	 end of turn."
//
// Sunburst is a keyword the engine reads off the resolving spell (ADR
// 0109 §11, #1552). The pump is an instant-speed ability with a counter
// cost.
func init() {
	Register(Spec{
		OracleID:            "3a8f382e-0bc6-4e79-87e2-9f99d907d234",
		Name:                "Sawtooth Thresher",
		Completeness:        CompletenessCaveats,
		Caveats:             []string{"With strict mana off, the game doesn't track which mana you spent, so it enters with no counters at all."},
		WantsDistinctColors: true,
		PrintedKeywords:     []string{game.KeywordSunburst},
		Activated: []ActivatedAbility{{
			Label:   "Remove two +1/+1 counters from this creature: It gets +4/+4 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    RemoveCountersFromThis(game.CounterPlusOne, 2),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return BoostUntilEOT{Target: item.SourceCardID, Power: 4, Toughness: 4}.Apply(NewContext(g, item))
			},
		}},
	})
}
