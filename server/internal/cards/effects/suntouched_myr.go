package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Suntouched Myr — Artifact Creature — Myr {3}, 0/0:
//
//	"Sunburst (This creature enters with a +1/+1 counter on it for each
//	 color of mana spent to cast it.)"
//
// Sunburst is a keyword the engine reads off the resolving spell (ADR
// 0109 §11, #1552). Cast off colourless mana it enters as a 0/0 and
// dies, as printed. Spec.WantsDistinctColors has the cast gate spread
// the payment across colours.
func init() {
	Register(Spec{
		OracleID:            "75d4f61a-8276-43ad-99e6-0d756a597e4b",
		Name:                "Suntouched Myr",
		Completeness:        CompletenessCaveats,
		Caveats:             []string{"With strict mana off, the game doesn't track which mana you spent, so it enters with no counters at all."},
		WantsDistinctColors: true,
		PrintedKeywords:     []string{game.KeywordSunburst},
	})
}
