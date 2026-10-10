package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "80c2784b-f27d-4a51-9380-171eb32b961d",
		Name:         "Inspired Tethermage",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouPutLoyaltyCountersOnAPlaneswalker("Inspired Tethermage — a +1/+1 counter on it",
				func(g *game.Game, item *game.StackItem) error {
					return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{6}: Empower Jace 2.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    ManaCost("{6}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return EmpowerJace{N: 2}.Apply(NewContext(g, item))
			},
		}},
	})
}
