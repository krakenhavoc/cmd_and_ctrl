package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Moon-Eating Dog — Creature — Dog {3}{U}, 3/3 (Global Series Jiang Yanggu & Mu Yanling, #5528):
//
//	"As long as you control a Yanling planeswalker, this creature has flying."
//
// Celestial Messenger's Yanling condition (SelfWhileYouControlA over a planeswalker with the Yanling subtype), granting flying.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a8e69725-cd57-4340-83b1-498bc0bd77b8",
		Name:         "Moon-Eating Dog",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.KeywordGrant(effects.SelfWhileYouControlA(effects.And(effects.Planeswalker(), effects.Subtype("Yanling"))), "flying")},
	})
}
