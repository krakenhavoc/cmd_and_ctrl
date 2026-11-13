package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dauthi Horror — Creature — Dauthi Horror {1}{B}, 2/1 (Tempest, #5628):
//
//	"Shadow (This creature can block or be blocked by only creatures
//	 with shadow.)
//	 This creature can't be blocked by white creatures."
//
// Shadow is a printed keyword (CR 702.28b). The second line is Stromkirk
// Noble's CantBeBlockedBy block rule (CR 509.1b) with a colour in place
// of a creature type: a blocker's colours are read live as blocks are
// declared, so a white creature with shadow can't block it, and a
// nonwhite one with shadow can. It is not protection from white: white
// sources still deal it damage and can target it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "07c20708-c305-46e7-a98e-c99f54aeafec",
		Name:            "Dauthi Horror",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"shadow"},
		BlockRules: []game.BlockRule{
			CantBeBlockedBy(OnSelf(), OfColor("W"), "white creatures"),
		},
	})
}
