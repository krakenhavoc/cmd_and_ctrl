package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kwende, Pride of Femeref — Legendary Creature — Human Knight {3}{W}, 2/2:
//
//	"Double strike
//	 Creatures you control with first strike have double strike."
//
// The grant reads first strike, which layer 6 adds and removes, so it is
// applied after every effect that would give one of your creatures first
// strike or take it away, whatever the timestamps (CR 613.8a, 613.8b;
// WithKeywordHave, #2663). That is the 2018-04-27 ruling: "If a creature
// you control gains first strike after Kwende has entered the
// battlefield, that creature also gains double strike." Having double
// strike is not having first strike, so Kwende's own double strike gives
// nothing to a creature that has only double strike, and a creature with
// both deals combat damage as one with double strike does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3adebbbf-39fe-4b42-b719-2f8edb009693",
		Name:            "Kwende, Pride of Femeref",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Static: []game.StaticAbility{
			WithKeywordHave(CreaturesYouControl, "first strike",
				"Creatures you control with first strike have double strike.", "double strike"),
		},
	})
}
