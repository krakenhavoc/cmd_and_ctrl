package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Beloved Princess — Creature — Human Noble {W}, 1/1 (Throne of Eldraine, #5516):
//
//	"Lifelink
//	 This creature can't be blocked by creatures with power 3 or greater."
//
// Exit Specialist's block rule over the blocker's power.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "682cbf2a-2723-4590-ae52-7c5d8b2573cf",
		Name:            "Beloved Princess",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		BlockRules: []game.BlockRule{
			CantBeBlockedBy(OnSelf(), PowerGE(3), "creatures with power 3 or greater"),
		},
	})
}
