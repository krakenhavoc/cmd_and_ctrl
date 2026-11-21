package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tattered Haunter — Creature — Spirit {1}{U}, 2/1 (Eldritch Moon, #5541):
//
//	"Flying
//	 This creature can block only creatures with flying."
//
// Wanderlight Spirit's two lines.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "754a72fd-e837-44f7-9bd7-5bf55ee8134c",
		Name:            "Tattered Haunter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		BlockRules: []game.BlockRule{
			effects.CantBlockAttackers(effects.OnSelf(), effects.OnMatching(effects.WithoutKeyword("flying")), "it can block only creatures with flying"),
		},
	})
}
