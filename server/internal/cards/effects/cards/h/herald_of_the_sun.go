package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Herald of the Sun — Creature — Angel {4}{W}{W}, 4/4 (Foundations
// Jumpstart, BNS batch 32):
//
//	"Flying
//	 {3}{W}: Put a +1/+1 counter on another target creature with flying."
//
// Benevolent Hydra's "another target creature" (the Herald cannot target
// itself; a second Herald is legal) with a flying clause read through the
// layered view at announcement and again at resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8331fcd2-b91d-45eb-8361-74aa73e4ff69",
		Name:            "Herald of the Sun",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{W}: Put a +1/+1 counter on another target creature with flying.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{3}{W}"),
			Targets: effects.Another(effects.TargetCreature("another target creature with flying", effects.HasKeyword("flying"))),
			Effect:  effects.PlusOneCounterOnTheTarget,
		}},
	})
}
