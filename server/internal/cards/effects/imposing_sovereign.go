package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Imposing Sovereign — Creature — Human Noble {1}{W}, 2/1 (Magic 2014, #5563):
//
//	"Creatures your opponents control enter tapped."
//
// Kinjalli's Sunwing's replacement on a 2/1: a creature an opponent controls
// enters tapped (CR 614.1c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4bba3663-7f3c-4589-88aa-cf0bfe36b171",
		Name:         "Imposing Sovereign",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b31OpponentsCreaturesEnterTapped("Imposing Sovereign: enters tapped"),
		},
	})
}
