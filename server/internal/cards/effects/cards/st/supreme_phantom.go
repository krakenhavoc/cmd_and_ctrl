package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Supreme Phantom — Creature — Spirit {1}{U}, 1/3:
//
//	"Flying
//	 Other Spirits you control get +1/+1."
//
// Legion Lieutenant's tribal anthem over Spirits.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9c5f4d02-eedb-4c6c-9f13-1b7a45382483",
		Name:            "Supreme Phantom",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Spirit"}, Others: true, YoursOnly: true}, 1, 1),
		},
	})
}
