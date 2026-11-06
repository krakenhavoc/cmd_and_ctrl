package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Offender at Large — Creature — Giant Rogue {4}{R}, 5/4 (Murders at
// Karlov Manor, #4469):
//
//	"Disguise {4}{R} (You may cast this card face down for {3} as a 2/2
//	 creature with ward {2}. Turn it face up any time for its disguise
//	 cost.)
//	 When this creature enters or is turned face up, up to one target
//	 creature gets +2/+0 until end of turn."
//
// Culvert Ambusher's one ability with two conditions (OnAny). Entering
// face down does not trigger it, because a face-down permanent has no
// abilities (CR 708.3); turning it face up does (CR 708.8). The pump is
// Daybreak Charger's, with "up to one" letting the controller pick none.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "c8c42851-1b1b-49dd-a97c-7eda57dadb13",
		Name:             "Offender at Large",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Disguise("{4}{R}")},
		Triggered: []game.TriggeredAbility{
			Targeting(
				OnAny([]game.EventKind{game.EventETB, game.EventTurnedFaceUp}, Self,
					"Offender at Large — up to one target creature gets +2/+0 until end of turn",
					boostTargetUntilEOT("Offender at Large — +2/+0", 2, 0)),
				TargetCreature("up to one target creature").WithCount(0, 1)),
		},
	})
}
