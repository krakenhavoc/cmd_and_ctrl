package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chasm Drake — Creature — Drake {4}{U}, 3/3 (Magic 2012, #5573):
//
//	"Flying
//	 Whenever this creature attacks, target creature you control gains
//	 flying until end of turn."
//
// Finch Formation's flying grant on an attack trigger instead of an enters
// one; the Drake may target itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e0f44bb8-f07b-416c-b767-87b9fe6461aa",
		Name:            "Chasm Drake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverThisAttacks("Chasm Drake — target creature you control gains flying until end of turn",
				effects.EbTargetCreatureGainsUntilEOT("flying", "Chasm Drake — flying until end of turn")),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}
