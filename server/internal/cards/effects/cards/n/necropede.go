package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Necropede — Artifact Creature — Phyrexian Insect {2}, 1/1 (Scars of
// Mirrodin, #5577):
//
//	"Infect
//	 When this creature dies, you may put a -1/-1 counter on target
//	 creature."
//
// Infect is the engine's damage tail (CR 702.90). The dies trigger is
// Hapatra's optional counter on a target creature, asked as it resolves
// (CR 603.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e5021798-1f3d-42ed-be88-e76349ef46c6",
		Name:            "Necropede",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"infect"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.WhenThisDies("Necropede — you may put a -1/-1 counter on target creature",
					effects.PutACounterOnTheTarget(game.CounterMinusOne)),
				effects.Harmful(effects.TargetCreature("target creature"))),
				"Necropede — put a -1/-1 counter on the target creature?"),
		},
	})
}
