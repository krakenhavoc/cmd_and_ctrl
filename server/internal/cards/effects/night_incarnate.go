package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Night Incarnate — Creature — Elemental {4}{B}, 3/4 (Commander 2018,
// #5526):
//
//	"Deathtouch
//	 When this creature leaves the battlefield, all creatures get -3/-3
//	 until end of turn.
//	 Evoke {3}{B} (You may cast this spell for its evoke cost. If you do,
//	 it's sacrificed when it enters.)"
//
// Evoke is Slithermuse's alternative cost: the creature really enters, the
// sacrifice is its own trigger, and the leaves trigger then sees the board.
// The leaves trigger fires on any departure (dies, exile, bounce). The
// -3/-3 is Rise of the Deathbringer's: every creature on the battlefield as
// the trigger resolves, locked in then (CR 611.2c), so a creature that
// enters afterwards is unaffected.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "abd1c415-8fd3-4852-92c6-4557453dfc41",
		Name:            "Night Incarnate",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Purpose:         game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 3}},
		AlternativeCosts: []game.AlternativeCost{
			Evoke("{3}{B}"),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisLeaves("Night Incarnate — all creatures get -3/-3 until end of turn",
				Do(BoostUntilEOT{Match: Creature(), Power: -3, Toughness: -3, Label: "Night Incarnate — -3/-3"})),
		},
	})
}
