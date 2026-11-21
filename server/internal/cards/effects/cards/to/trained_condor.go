package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trained Condor — Creature — Bird {2}{U}, 2/1 (Magic 2014, #5563):
//
//	"Flying
//	 Whenever this creature attacks, another target creature you control gains flying until end of turn."
//
// Starling, Aerial Ally's grant, on attack. "Another" is object identity
// (Another), so the Condor can't pick itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "99742ec7-7b67-4301-9307-3c2068046e33",
		Name:            "Trained Condor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverThisAttacks("Trained Condor — another target creature you control gains flying until end of turn",
				effects.EbTargetCreatureGainsUntilEOT("flying", "Trained Condor — flying until end of turn")),
				effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))),
		},
	})
}
