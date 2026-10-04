package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Howlgeist — Creature — Spirit Wolf {5}{G}, 4/2:
//
//	"Creatures with power less than this creature's power can't block it.
//	 Undying"
//
// The comparison is made as blockers are declared, against Howlgeist's
// power then, so the +1/+1 counter undying puts on it raises the bar.
// Undying is PrintedKeywords (#2075).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5a00eb94-6c0f-4f41-8c16-fc4e2f7cb9aa",
		Name:            "Howlgeist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordUndying},
		BlockRules: []game.BlockRule{
			CantBlockAttackers(PowerLessThanSource(), OnSelf(),
				"creatures with power less than Howlgeist's can't block it"),
		},
	})
}
