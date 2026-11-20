package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "99742ec7-7b67-4301-9307-3c2068046e33",
		Name:            "Trained Condor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisAttacks("Trained Condor — another target creature you control gains flying until end of turn",
				ebTargetCreatureGainsUntilEOT("flying", "Trained Condor — flying until end of turn")),
				Another(TargetCreature("another target creature you control", YouControl()))),
		},
	})
}
