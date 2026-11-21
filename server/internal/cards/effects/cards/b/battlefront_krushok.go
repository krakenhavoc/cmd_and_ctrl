package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Battlefront Krushok — Creature — Beast {4}{G}, 3/4 (Ravnica: Clue
// Edition):
//
//	"This creature can't be blocked by more than one creature.
//	 Each creature you control with a +1/+1 counter on it can't be blocked
//	 by more than one creature."
//
// Two block-count rules (MaxBlockers, CR 509.1b): one on the Krushok, one on
// every creature its controller controls that has a +1/+1 counter, read
// live as each block is declared.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6a003988-f126-4309-9166-50f3fbf0f30c",
		Name:         "Battlefront Krushok",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.MaxBlockers(effects.OnSelf(), 1),
			effects.MaxBlockers(effects.OnMatching(effects.And(effects.Creature(), effects.YouControl(), effects.VowB09HasPlusOneCounter())), 1),
		},
	})
}
