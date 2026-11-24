package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wren's Run Hydra — Creature — Hydra {X}{G}, 0/0 (Modern Horizons 2,
// #5501):
//
//	"Reach
//	 This creature enters with X +1/+1 counters on it.
//	 Reinforce X—{X}{G}{G} ({X}{G}{G}, Discard this card: Put X +1/+1
//	 counters on target creature.)"
//
// The entry counters are the X of the cast (XCounters, the Benevolent
// Hydra shape). Reinforce X is a hand ability whose {X} is announced as it
// is activated (S82OldB37ReinforceX).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "a2196bc3-2fe6-49d4-8c79-d6beb3d6c5d8",
		Name:                       "Wren's Run Hydra",
		Completeness:               effects.CompletenessFull,
		XMatters:                   true,
		PrintedKeywords:            []string{"reach"},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.XCounters(game.CounterPlusOne)},
		Activated: []effects.ActivatedAbility{
			effects.S82OldB37ReinforceX("{X}{G}{G}", "Reinforce X—{X}{G}{G} ({X}{G}{G}, Discard this card: Put X +1/+1 counters on target creature.)"),
		},
	})
}
