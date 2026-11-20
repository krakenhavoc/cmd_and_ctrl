package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wary Thespian — Creature — Cat Druid {1}{G}, 3/1 (Foundations, #4465):
//
//	"When this creature enters or dies, surveil 1."
//
// Thawbringer's enters-or-dies trigger on a smaller body: one ability
// with two conditions, so it surveils once as it arrives and once more
// as it dies.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "448cc829-1596-4aec-b180-c0c1e0566f6b",
		Name:         "Wary Thespian",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersOrDies("Wary Thespian — surveil 1", effects.Do(effects.Surveil{N: 1})),
		},
	})
}
