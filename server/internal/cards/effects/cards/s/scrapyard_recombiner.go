package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scrapyard Recombiner — Artifact Creature — Construct {3}, 0/0 (Modern
// Horizons):
//
//	"Modular 2 (This creature enters with two +1/+1 counters on it. When it
//	 dies, you may put its +1/+1 counters on target artifact creature.)
//	 {T}, Sacrifice an artifact: Search your library for a Construct card,
//	 reveal it, put it into your hand, then shuffle."
//
// Modular is the engine's keyword (CR 702.43a, Arcbound Reclaimer's). The
// ability's cost is the tap symbol, so a Recombiner with summoning sickness
// cannot pay it (CR 302.6), and a sacrificed artifact (the Recombiner itself
// qualifies) leaves as part of the cost. The search finds a card with the
// Construct subtype as it is in the library, reveals it, puts it into hand
// and shuffles, and may find nothing (CR 701.23b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ecf09f2d-a886-4ec8-8ff6-fe15d87b1375",
		Name:            "Scrapyard Recombiner",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"modular 2"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice an artifact: Search your library for a Construct card, reveal it, put it into your hand, then shuffle.",
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeN(1, "an artifact", effects.Artifact())),
			Purpose: game.Purpose{Tutors: 1, Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: func(c game.Card) bool { return c.HasSubtypeWord("Construct") },
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Scrapyard Recombiner — a Construct card",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
