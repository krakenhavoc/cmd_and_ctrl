package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Court of Bounty — Enchantment {2}{G}{G}:
//
//	"When this enchantment enters, you become the monarch.
//	 At the beginning of your upkeep, you may put a land card from your
//	 hand onto the battlefield. If you're the monarch, instead you may
//	 put a creature or land card from your hand onto the battlefield."
//
// The green Court (#1722): the put_from_hand.go primitive with its
// filter chosen at resolution, since the monarch clause is part of the
// effect. The put is a "you may" (Optional, so declining is a real
// answer) and is not a land drop — putting a land onto the battlefield
// is not playing one (CR 305.2). A player with no qualifying card is
// not prompted.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e55d9377-f89a-41e5-a094-730d6f24caf0",
		Name:         "Court of Bounty",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouBecomeTheMonarch("Court of Bounty"),
			AtYourUpkeep("Court of Bounty — put a land, or a creature or land if you're the monarch, from your hand onto the battlefield",
				courtOfBountyUpkeep),
		},
	})
}

// courtOfBountyUpkeep offers the land, or for the monarch the creature
// or land.
func courtOfBountyUpkeep(g *game.Game, item *game.StackItem) error {
	put := PutFromHandOntoBattlefield{
		Match:    Land(),
		Optional: true,
		Label:    "Court of Bounty — you may put a land card from your hand onto the battlefield",
	}
	if YoureTheMonarch(g, item.Controller) {
		put.Match = Or(Creature(), Land())
		put.Label = "Court of Bounty — you may put a creature or land card from your hand onto the battlefield"
	}
	return put.Apply(NewContext(g, item))
}
