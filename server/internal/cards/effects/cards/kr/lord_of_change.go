package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lord of Change — Creature — Demon {6}{U}, 6/6 (Warhammer 40,000 Commander,
// S82 GAP batch 195, #5948):
//
//	"Flying, ward {3}
//	 Architect of Deception — When this creature enters, draw three cards."
//
// Flying, the ward trigger and an enters draw of three.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b575fb16-c3c3-4cad-857f-122257a4b32f",
		Name:            "Lord of Change",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{3}"), "Lord of Change — ward {3}"),
			effects.WhenThisEnters("Lord of Change — draw three cards", effects.Do(effects.DrawCards{N: 3})),
		},
	})
}
