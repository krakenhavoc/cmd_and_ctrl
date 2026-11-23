package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rustvine Cultivator — Creature — Phyrexian Elf Druid {G}, 1/2 (Phyrexia:
// All Will Be One):
//
//	"{T}: Put an oil counter on this creature.
//	 {T}, Remove an oil counter from this creature: Untap target land."
//
// It enters with no counters. Both abilities tap it, so one use a turn
// unless it is untapped; the land may be anyone's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7a49731b-bd95-4f0a-ac47-1aa318c12fec",
		Name:         "Rustvine Cultivator",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}: Put an oil counter on this creature.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.TapCost(),
				Effect:  effects.Do(effects.CounterOnThis{Kind: effects.OilCounter, N: 1}),
			},
			{
				Label:   "{T}, Remove an oil counter from this creature: Untap target land.",
				Cost:    effects.Plus(effects.TapCost(), effects.RemoveCountersFromThis(effects.OilCounter, 1)),
				Targets: effects.TargetPermanent("target land", effects.Land()),
				Effect:  effects.UntapEachLegalTarget,
			},
		},
	})
}
