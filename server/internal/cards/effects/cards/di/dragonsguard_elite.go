package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dragonsguard Elite — Creature — Human Druid {1}{G}, 2/2 (Strixhaven:
// School of Mages, #5502):
//
//	"Magecraft — Whenever you cast or copy an instant or sorcery spell, put a
//	 +1/+1 counter on this creature.
//	 {4}{G}{G}: Double the number of +1/+1 counters on this creature."
//
// The magecraft is Baxter's counter on the source. The doubling is Solarion's
// (CR 701.10e): as many +1/+1 counters as it has, read as the ability
// resolves, so a counter doubler applies to the ones put on.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9bc966b3-1b6e-4bea-ab1d-321bc03d2bc7",
		Name:         "Dragonsguard Elite",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Magecraft("Dragonsguard Elite — put a +1/+1 counter on it (magecraft)",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{G}{G}: Double the number of +1/+1 counters on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{4}{G}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.B23DoublePlusOneCountersOn(effects.NewContext(g, item), item.SourceCardID)
			},
		}},
	})
}
