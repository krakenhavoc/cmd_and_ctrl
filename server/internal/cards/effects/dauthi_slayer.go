package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dauthi Slayer — Creature — Dauthi Soldier {B}{B}, 2/2 (Tempest, #5628):
//
//	"Shadow (This creature can block or be blocked by only creatures
//	 with shadow.)
//	 This creature attacks each combat if able."
//
// Shadow is a printed keyword (CR 702.28b). The attack requirement is
// Flamewake Phoenix's AttacksEachCombat static, checked as attackers are
// declared (CR 508.1d): its controller can't finish the declaration with
// the Slayer at home while it is untapped and able to attack, and a
// summoning-sick or tapped Slayer is not required to.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c1d0537b-954f-4d3b-b9fe-1dd7ffaa1197",
		Name:            "Dauthi Slayer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"shadow"},
		Static:          []game.StaticAbility{AttacksEachCombat()},
	})
}
