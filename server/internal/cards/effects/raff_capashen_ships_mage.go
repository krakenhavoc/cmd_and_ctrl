package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Raff Capashen, Ship's Mage — Legendary Creature — Human Wizard {2}{W}{U},
// 3/3:
//
//	"Flash
//	 Flying
//	 You may cast historic spells as though they had flash. (Artifacts,
//	 legendaries, and Sagas are historic.)"
//
// Flash and flying are printed keywords. The third ability is Shimmer
// Myr's battlefield-derived cast timing (CastKindAsThoughFlash) over the
// permission filter's historic test (an artifact, a legendary or a Saga,
// CR 700.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "13044ca1-76a3-43b2-a8e4-3bbc692c30c5",
		Name:            "Raff Capashen, Ship's Mage",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		CastTimings: []game.CastTimingRule{
			CastKindAsThoughFlash(game.PermissionFilter{HistoricOnly: true},
				"You may cast historic spells as though they had flash."),
		},
	})
}
