package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Solemn Recruit — Creature — Dwarf Warrior {1}{W}{W}, 2/2 (Aether Revolt,
// #5536):
//
//	"Double strike
//	 Revolt — At the beginning of your end step, if a permanent left the
//	 battlefield under your control this turn, put a +1/+1 counter on this
//	 creature."
//
// Shortcut to Mushrooms' end-step revolt trigger (revolt.go), with the
// counter on the Recruit itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cf747d6e-05ae-484d-9e72-1d0fb71fd4ff",
		Name:            "Solemn Recruit",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStepIfRevolt("Solemn Recruit — put a +1/+1 counter on this creature",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
