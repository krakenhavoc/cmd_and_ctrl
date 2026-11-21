package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wary Watchdog — Creature — Dog {1}{G}, 3/1 (Duskmourn: House of
// Horror, #4466):
//
//	"When this creature enters or dies, surveil 1."
//
// Wary Thespian's enters-or-dies trigger: one ability with two
// conditions, so it surveils once as it arrives and once more as it
// dies.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7e063f60-876a-48fe-8bc6-9888180176d1",
		Name:         "Wary Watchdog",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersOrDies("Wary Watchdog — surveil 1", effects.Do(effects.Surveil{N: 1})),
		},
	})
}
