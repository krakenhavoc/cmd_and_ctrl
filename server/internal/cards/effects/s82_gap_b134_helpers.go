package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_gap_b134_helpers.go — shared pieces of the Dominaria Remastered GAP
// batch 134 cards (#5948, S82).

// s82GapB134ReturnSourceToHandBody is the delayed trigger of Wall of
// Junk: return the object named in Params.Object to its owner's hand, only
// if it is still that object and still on the battlefield. An on-disk key.
var s82GapB134ReturnSourceToHandBody game.BodyRef

func init() {
	s82GapB134ReturnSourceToHandBody = game.SimpleDelayedBody("s82-gap-b134/return-source-to-hand", s82GapB134ReturnSourceToHand)
}

func s82GapB134ReturnSourceToHand(g *game.Game, item *game.StackItem) error {
	obj := item.Params.Object
	if !onBattlefield(g, obj.ID) || g.ObjectEpochForEffect(obj.ID) != obj.Epoch {
		return nil
	}
	return BounceToHand{Target: obj.ID}.Apply(NewContext(g, item))
}
