package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stratozeppelid — Creature — Beast {4}{U}, 4/4 (Guildpact, #5599):
//
//	"Flying
//	 This creature can block only creatures with flying."
//
// Cloud Spirit's block rule on a bigger body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f20640db-c5d4-4e00-9903-a33ff4f7820d",
		Name:            "Stratozeppelid",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		BlockRules: []game.BlockRule{
			effects.CantBlockAttackers(effects.OnSelf(), effects.OnMatching(effects.WithoutKeyword("flying")), "it can block only creatures with flying"),
		},
	})
}
