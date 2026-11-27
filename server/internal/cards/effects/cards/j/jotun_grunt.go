package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jötun Grunt — Creature — Giant Soldier {1}{W}, 4/4 (Commander 2011, S82 older
// sets batch 454, #5574):
//
//	"Cumulative upkeep—Put two cards from a single graveyard on the bottom of
//	 their owner's library."
//
// The cumulative-upkeep payment of #6121: each age counter asks for two cards
// from one graveyard (CR 702.24a, each counter's choice made separately), all
// paid together or not at all; unpaid, the Grunt is sacrificed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "43fbfeec-bcaf-48b8-befe-b7346fec5a3a",
		Name:         "Jötun Grunt",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.CumulativeUpkeepPaying("Jötun Grunt — cumulative upkeep: put two cards from a single graveyard on the bottom of their owner's library",
				effects.GraveyardToLibraryBottomPayment(2, 2), "card", "cards"),
		},
	})
}
