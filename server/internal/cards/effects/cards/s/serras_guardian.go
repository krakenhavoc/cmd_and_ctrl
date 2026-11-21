package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Serra's Guardian — Creature — Angel {4}{W}{W}, 5/5 (Core Set 2019, #5527):
//
//	"Flying (This creature can't be blocked except by creatures with flying or reach.)
//	Vigilance (Attacking doesn't cause this creature to tap.)
//	Other creatures you control have vigilance."
//
// Aang, Air Nomad's text: its own flying and vigilance, and a layer-6
// grant of vigilance to every other creature you control.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c50cd7e7-8d53-44af-ac96-3e104587bdb3",
		Name:            "Serra's Guardian",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Static: []game.StaticAbility{
			effects.KeywordGrant(effects.OtherCreaturesYouControl, "vigilance"),
		},
	})
}
