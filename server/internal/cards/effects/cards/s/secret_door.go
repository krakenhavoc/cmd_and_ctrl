package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Secret Door — Artifact Creature — Wall {U}, 0/4:
//
//	"Defender
//	 {4}{U}: Venture into the dungeon. Activate only as a sorcery."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a53baf90-8d0c-4900-9f24-893ad31628ec",
		Name:            "Secret Door",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label:        "{4}{U}: Venture into the dungeon. Activate only as a sorcery.",
			Cost:         effects.ManaCost("{4}{U}"),
			SorcerySpeed: true,
			Effect:       effects.Venture,
		}},
	})
}
