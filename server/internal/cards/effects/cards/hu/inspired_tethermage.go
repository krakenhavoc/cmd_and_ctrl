package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inspired Tethermage — Creature — Elf Warrior {2}{G}, 3/2 (Reality
// Fracture, tracker #2795):
//
//	"Whenever you put one or more loyalty counters on a planeswalker,
//	 put a +1/+1 counter on this creature.
//	 {6}: Empower Jace 2."
//
// The trigger is ADR 0140's one-per-placement loyalty watcher; a Jace
// token this creature's own {6} ability empowers counts, so the ability
// grows it. The activation is the keyword action (ADR 0139).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "80c2784b-f27d-4a51-9380-171eb32b961d",
		Name:         "Inspired Tethermage",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouPutLoyaltyCountersOnAPlaneswalker("Inspired Tethermage — a +1/+1 counter on it",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{6}: Empower Jace 2.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{6}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.EmpowerJace{N: 2}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
