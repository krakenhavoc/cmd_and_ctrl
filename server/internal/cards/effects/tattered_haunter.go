package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tattered Haunter — Creature — Spirit {1}{U}, 2/1 (Eldritch Moon, #5541):
//
//	"Flying
//	 This creature can block only creatures with flying."
//
// Wanderlight Spirit's two lines.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "754a72fd-e837-44f7-9bd7-5bf55ee8134c",
		Name:            "Tattered Haunter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		BlockRules: []game.BlockRule{
			CantBlockAttackers(OnSelf(), OnMatching(WithoutKeyword("flying")), "it can block only creatures with flying"),
		},
	})
}
