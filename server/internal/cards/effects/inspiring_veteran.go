package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Inspiring Veteran — Creature — Human Knight {R}{W}, 2/2 (Throne of Eldraine, #5516):
//
//	"Other Knights you control get +1/+1."
//
// Marshal of Zhalfir's lord clause (TribalAnthem).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "aa1a63dd-acb1-465f-8970-667b8d7c57c9",
		Name:         "Inspiring Veteran",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Knight"}, Others: true, YoursOnly: true}, 1, 1),
		},
	})
}
