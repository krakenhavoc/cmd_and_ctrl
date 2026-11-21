package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thawbringer — Creature — Insect Scout {2}{G}, 4/2 (Edge of Eternities,
// #4461):
//
//	"When this creature enters or dies, surveil 1."
//
// Crow of Dark Tidings' enters-or-dies trigger (WhenThisEntersOrDies)
// with a surveil 1: one ability with two conditions, so it surveils once
// as it arrives and once more as it dies.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0863c4bc-1e3e-4ff9-8877-88d4021c85e5",
		Name:         "Thawbringer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersOrDies("Thawbringer — surveil 1", effects.Do(effects.Surveil{N: 1})),
		},
	})
}
