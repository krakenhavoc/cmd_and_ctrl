package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Living Tsunami — Creature — Elemental {2}{U}{U}, 4/4 (Zendikar, #5583):
//
//	"Flying
//	 At the beginning of your upkeep, sacrifice this creature unless you
//	 return a land you control to its owner's hand."
//
// The payment is asked as the trigger resolves (the Visions karoos'
// choice, with any one land): declining, or having no land, sacrifices it,
// and it stays only if the land really reached a hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7260657a-c1cd-4081-8ab5-46d604dccca8",
		Name:            "Living Tsunami",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{effects.S82OldB498SacrificeUnlessYouReturnALandAtUpkeep("Living Tsunami")},
	})
}
