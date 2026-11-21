package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Enter the Unknown — Sorcery {G} (Rivals of Ixalan, #5531):
//
//	"Target creature you control explores. (Reveal the top card of your
//	 library. Put that card into your hand if it's a land. Otherwise, put
//	 a +1/+1 counter on the creature, then put the card back or put it
//	 into your graveyard.)
//	 You may play an additional land this turn."
//
// The Map token's explore (CR 701.44a) and Explore's extra land drop. The
// land drop is granted first because the explore may stop to ask where
// the revealed card goes; nothing can be played in between, so the order
// is not visible. A target that left in response makes the whole spell
// do nothing (CR 608.2b), land drop included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4b031bb3-4bbc-4158-8485-b3fc42c6bd50",
		Name:         "Enter the Unknown",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{ExtraLandDrops: 1},
		Targets:      effects.TargetCreatureYouControl(),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			ctx.Game.GrantAdditionalLandPlayForEffect(item.Controller, 1)
			return effects.TargetCreatureYouControlExplores(ctx.Game, item)
		},
	})
}
