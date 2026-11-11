package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eccentric Farmer — Creature — Human Peasant {2}{G}, 2/3 (Outlaws of Thunder
// Junction Commander, #5368):
//
//	"When this creature enters, mill three cards, then you may return a land
//	 card from your graveyard to your hand."
//
// Pothole Mole's sentence: the land is chosen after the mill (CR 608.2c), so
// a land just milled is a legal pick, and "you may" lets the controller take
// none. One shared body (cmdr2B15MillThenMayReturnLandTrigger).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b51a852c-ccde-4d23-a3f3-a002f7eacdab",
		Name:         "Eccentric Farmer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			cmdr2B15MillThenMayReturnLandTrigger("Eccentric Farmer", 3,
				"mill three cards, then you may return a land card from your graveyard to your hand"),
		},
	})
}
