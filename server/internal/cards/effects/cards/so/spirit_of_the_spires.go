package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spirit of the Spires — Creature — Spirit {3}{W}, 2/4 (Ravnica Allegiance,
// #5521):
//
//	"Flying
//	 Other creatures you control with flying get +0/+1."
//
// Flying is printed data (CR 702.9b); the +0/+1 is Sprite Noble's layer-7c
// anthem over the other creatures you control with flying.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bf10ce45-c6d6-461e-ab46-06b4205d31fb",
		Name:            "Spirit of the Spires",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{effects.B16Anthem(effects.B29OtherCreaturesYouControlWithFlying, 0, 1)},
	})
}
