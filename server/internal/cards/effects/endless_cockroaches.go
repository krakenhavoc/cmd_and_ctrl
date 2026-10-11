package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Endless Cockroaches — Creature — Insect {1}{B}{B}, 1/1:
//
//	"When this creature dies, return it to its owner's hand."
//
// A graveyard-only leaves-the-battlefield trigger (WhenThisDies), so a
// Cockroaches that was exiled or bounced does not set it off. "It" is
// the object that died: returnSelfFromGraveyardToHand moves the card
// only while it is still the graveyard object the trigger stamped, so
// one that was exiled, reanimated or otherwise left the graveyard
// before the trigger resolved stays where it is (CR 400.7). It goes to
// its OWNER's hand, whoever controlled it when it died. A token copy
// ceased to exist in the graveyard (CR 111.7) and returns nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9f31dbb1-c350-46f8-bd1d-9f23a073d2f1",
		Name:         "Endless Cockroaches",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Endless Cockroaches — return it to its owner's hand", returnSelfFromGraveyardToHand),
		},
	})
}
