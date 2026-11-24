package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spell Satchel — Artifact {2} (Strixhaven: School of Mages, #5502):
//
//	"Magecraft — Whenever you cast or copy an instant or sorcery spell, put a
//	 book counter on this artifact.
//	 {T}, Remove a book counter from this artifact: Add {C}.
//	 {3}, {T}, Remove three book counters from this artifact: Draw a card."
//
// The mana ability takes its counter as part of the cost (Component Pouch's
// shape), so a Satchel with none cannot be tapped for mana; the draw is a
// stack ability with the three counters in its cost (Arcane Spyglass).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "912c5e71-554c-4e31-8831-018760752eb1",
		Name:         "Spell Satchel",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Magecraft("Spell Satchel — put a book counter on it (magecraft)",
				effects.Do(effects.CounterOnThis{Kind: "book", N: 1})),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost: effects.ManaAbilityCost{
				Tap:            true,
				RemoveCounters: effects.RemoveCountersFromThis("book", 1).RemoveCounters,
			},
			Produced: "{C}",
			Label:    "{T}, Remove a book counter from this artifact: Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}, Remove three book counters from this artifact: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost(), effects.RemoveCountersFromThis("book", 3)),
			Effect:  effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
