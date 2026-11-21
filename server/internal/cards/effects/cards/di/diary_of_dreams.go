package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Diary of Dreams — Artifact — Book {2} (Secrets of Strixhaven, #4456):
//
//	"Whenever you cast an instant or sorcery spell, put a page counter on
//	 this artifact.
//	 {5}, {T}: Draw a card. This ability costs {1} less to activate for
//	 each page counter on this artifact."
//
// The counter trigger is one cast trigger (CounterOnThis, the one-sentence
// footprint). The cost clause is the ability's own cost modifier
// (ActivatedAbility.CostModifiers; Mariposa Military Base's shape), counted
// as the cost is determined (CR 602.2b, 601.2f) from the page counters on
// the artifact: {1} less per counter, never below {0} generic, while the {T}
// is still paid.
//
// No simplification.
func init() {
	const page = "page"
	effects.Register(effects.Spec{
		OracleID:     "98b43c61-4563-4a29-8cf0-fb8b34f337ce",
		Name:         "Diary of Dreams",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Or(effects.Instant(), effects.Sorcery()),
				"Diary of Dreams — put a page counter on this artifact",
				effects.Do(effects.CounterOnThis{Kind: page, N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}, {T}: Draw a card. This ability costs {1} less to activate for each page counter on this artifact.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{5}"), effects.TapCost()),
			CostModifiers: []game.CostModifier{
				effects.CostsLessEach(effects.CountersOnSource(page),
					"This ability costs {1} less to activate for each page counter on this artifact."),
			},
			Effect: effects.B36DrawOne,
		}},
	})
}
