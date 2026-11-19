package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dreamdrinker Vampire — Creature — Vampire {1}{B}, 2/1 (Modern Horizons 3):
//
//	"Lifelink
//	 {1}{B}: Adapt 1.
//	 Whenever one or more +1/+1 counters are put on this creature, it gains
//	 menace until end of turn."
//
// One trigger per placement event (CR 603.2c), so a doubled adapt still
// gives one menace grant. Adapt checks for counters as it resolves (CR 701.46a).
func init() {
	const label = "Dreamdrinker Vampire — gains menace until end of turn"
	Register(Spec{
		OracleID:        "5b4e1163-09aa-4597-9a41-e1e384975056",
		Name:            "Dreamdrinker Vampire",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}: Adapt 1.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{1}{B}"),
			Effect:  func(g *game.Game, item *game.StackItem) error { return adaptN(g, item, 1) },
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventCounterPlaced, gapB49CountersPlacedOnThis, label,
				func(g *game.Game, item *game.StackItem) error {
					return GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"menace"}, Label: label}.Apply(NewContext(g, item))
				}),
		},
	})
}
