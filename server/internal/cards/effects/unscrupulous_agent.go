package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unscrupulous Agent — Creature — Elf Detective {1}{B}, 1/1 (Murders at
// Karlov Manor, #4469):
//
//	"When this creature enters, target opponent exiles a card from their
//	 hand."
//
// Ruthless Negotiation's pick (std_b40_helpers.go): the target opponent
// chooses the card from their own hand, a floor of one, and an empty hand
// exiles nothing (CR 609.3). A target that has become illegal makes the
// ability do nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3e7879d5-62ea-4c9a-9fc0-659d70f3a8e1",
		Name:         "Unscrupulous Agent",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Unscrupulous Agent — target opponent exiles a card from their hand",
					func(g *game.Game, item *game.StackItem) error {
						return targetPlayerExilesFromHand(NewContext(g, item), "Unscrupulous Agent — exile a card from your hand")
					}),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}
