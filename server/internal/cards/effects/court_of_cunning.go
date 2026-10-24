package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Court of Cunning — Enchantment {1}{U}{U}:
//
//	"When this enchantment enters, you become the monarch.
//	 At the beginning of your upkeep, any number of target players each
//	 mill two cards. If you're the monarch, each of those players mills
//	 ten cards instead."
//
// The blue Court (#1722). "Any number of target players" is the (0, 0)
// count Singularity Rupture declares — zero is a legal choice, and
// yourself is a legal pick (a graveyard deck mills itself). Each target
// is re-checked at resolution (CR 608.2b); the count is read then too,
// since the monarch clause is part of the effect.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2ab88c99-aaa0-4a91-9225-0bbfba04b6bc",
		Name:         "Court of Cunning",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouBecomeTheMonarch("Court of Cunning"),
			Targeting(AtYourUpkeep("Court of Cunning — each target player mills two, or ten if you're the monarch", courtOfCunningUpkeep),
				TargetPlayer("any number of target players").WithCount(0, 0)),
		},
	})
}

// courtOfCunningUpkeep mills each chosen player two, or ten for the
// monarch.
func courtOfCunningUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	n := 2
	if YoureTheMonarch(g, item.Controller) {
		n = 10
	}
	var players []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			players = append(players, t.ID)
		}
	}
	// Each chosen player mills in APNAP order, one prompt at a time (CR 101.4).
	return millEachPlayerInTurn(ctx, players, constantMill(n), nil)
}
