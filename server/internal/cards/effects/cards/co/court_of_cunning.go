package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "2ab88c99-aaa0-4a91-9225-0bbfba04b6bc",
		Name:         "Court of Cunning",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouBecomeTheMonarch("Court of Cunning"),
			effects.Targeting(effects.AtYourUpkeep("Court of Cunning — each target player mills two, or ten if you're the monarch", CourtOfCunningUpkeep),
				effects.TargetPlayer("any number of target players").WithCount(0, 0)),
		},
	})
}

// courtOfCunningUpkeep mills each chosen player two, or ten for the
// monarch.
func CourtOfCunningUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	n := 2
	if effects.YoureTheMonarch(g, item.Controller) {
		n = 10
	}
	var players []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			players = append(players, t.ID)
		}
	}
	// Each chosen player mills in APNAP order, one prompt at a time (CR 101.4).
	return effects.MillEachPlayerInTurn(ctx, players, effects.ConstantMill(n), game.KeyedTail{})
}
