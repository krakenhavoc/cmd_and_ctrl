package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kongming, "Sleeping Dragon" — Legendary Creature — Human Advisor {2}{W}{W},
// 2/2 (Commander 2013, S82 OLD b387):
//
//	"Other creatures you control get +1/+1."
//
// Benalish Marshal's anthem: a layer 7c boost over the other creatures you
// control; Kongming itself is not pumped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "21e9e1a9-5d6d-473e-adab-6a1e8e2b0ebd",
		Name:         "Kongming, \"Sleeping Dragon\"",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.B16Anthem(effects.OtherCreaturesYouControl, 1, 1)},
	})
}
