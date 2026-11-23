package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lab Rats — Sorcery {B} (Stronghold, #5627):
//
//	"Buyback {4} (You may pay an additional {4} as you cast this spell.
//	 If you do, put this card into your hand as it resolves.)
//	 Create a 1/1 black Rat creature token."
//
// The token table's "1/1 black Rat" row (Ogre Slumlord's), created
// through ResolveSteps as Acorn Harvest does, with Capsize's buyback
// (CR 702.27a). The spell has no target, so it always resolves, and with
// buyback paid it goes back to its owner's hand every time.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "c53d6d3e-64e3-48b9-afab-3bde3164888f",
		Name:          "Lab Rats",
		Completeness:  effects.CompletenessFull,
		Purpose:       game.Purpose{Tokens: 1, CreatureTokens: 1},
		OptionalCosts: []game.AdditionalCost{effects.Buyback("{4}")},
		OnResolve:     effects.ResolveSteps(effects.CreateToken{Template: effects.TokenCard("1/1 black Rat"), N: 1}),
	})
}
