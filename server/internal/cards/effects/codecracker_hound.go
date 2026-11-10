package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Codecracker Hound — Creature — Dog {2}{U}, 2/1 (Edge of Eternities,
// #4461):
//
//	"When this creature enters, look at the top two cards of your
//	 library. Put one into your hand and the other into your graveyard.
//	 Warp {2}{U}"
//
// Faerie Snoop's look-and-take over the top two (a mandatory pick for the
// hand, the other into the graveyard) on an enters trigger, plus the
// shared warp cost (CR 702.185a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "ae791b61-5825-4987-a8c2-80813758d3f7",
		Name:             "Codecracker Hound",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Warp("{2}{U}")},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Codecracker Hound — look at the top two cards of your library, one to hand and the other to your graveyard",
				codecrackerHoundLook),
		},
	})
}

func codecrackerHoundLook(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player := ctx.Controller()
	return TakeFromLibraryToHand{
		Player: player,
		Cards:  g.LookAtTopOfLibraryForEffect(player, 2),
		Max:    1,
		Label:  "Codecracker Hound — put one of them into your hand and the other into your graveyard",
		Key:    takeRestGraveyard,
	}.Apply(ctx)
}
