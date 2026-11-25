package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Heavenly Qilin — Creature — Kirin {2}{W}, 2/2 (Global Series Jiang Yanggu & Mu Yanling, #5528):
//
//	"Flying"
//	"Whenever this creature attacks, another target creature you control gains flying until end of turn."
//
// The target is chosen as the trigger goes on the stack and re-checked on resolution (CR 608.2b), as for Appa, Aang's Companion.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cbe1c80a-db82-460a-b6e9-d0e429628aa1",
		Name:            "Heavenly Qilin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverThisAttacks("Heavenly Qilin — another target creature you control gains flying until end of turn", effects.FirstTargetGainsUntilEndOfTurn(0, "Heavenly Qilin — another target creature you control gains flying until end of turn", "flying")),
				effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))),
		},
	})
}
