package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stratozeppelid — Creature — Beast {4}{U}, 4/4 (Guildpact, #5599):
//
//	"Flying
//	 This creature can block only creatures with flying."
//
// Cloud Spirit's block rule on a bigger body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f20640db-c5d4-4e00-9903-a33ff4f7820d",
		Name:            "Stratozeppelid",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		BlockRules: []game.BlockRule{
			CantBlockAttackers(OnSelf(), OnMatching(WithoutKeyword("flying")), "it can block only creatures with flying"),
		},
	})
}
