package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Exit Specialist — Creature — Human Detective {1}{U}, 2/1 (Murders at
// Karlov Manor, #4469):
//
//	"This creature can't be blocked by creatures with power 3 or greater.
//	 Disguise {1}{U} (You may cast this card face down for {3} as a 2/2
//	 creature with ward {2}. Turn it face up any time for its disguise
//	 cost.)
//	 When this creature is turned face up, return another target creature
//	 to its owner's hand."
//
// The block rule is Bilbo's (CantBeBlockedBy with PowerGE(3)); face down
// the creature has no text (CR 708.2a), so a face-down Specialist can be
// blocked by anything. The face-up trigger (CR 708.8) is Forum Familiar's
// bounce, aimed at "another" creature by object identity (Another), and
// a target that left in response is skipped (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "79d2efb2-145d-4377-a4e2-832ff2cb5677",
		Name:         "Exit Specialist",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{
			CantBeBlockedBy(OnSelf(), PowerGE(3), "creatures with power 3 or greater"),
		},
		AlternativeCosts: []game.AlternativeCost{Disguise("{1}{U}")},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisIsTurnedFaceUp("Exit Specialist — return another target creature to its owner's hand",
					bounceTheTargetToHand),
				Another(TargetCreature("another target creature"))),
		},
	})
}
