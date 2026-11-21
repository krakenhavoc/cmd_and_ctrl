package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Advanced Hoverguard — Creature — Drone {3}{U}, 2/2 (Fifth Dawn, #5604):
//
//	"Flying
//	 {U}: This creature gains shroud until end of turn. (It can't be the
//	 target of spells or abilities.)"
//
// Killer Whale's self-grant with shroud in place of flying. Shroud is
// the keyword the targeting choke point already honours, so a spell or
// ability already aimed at the Hoverguard is left with an illegal target
// when it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "65dc61a4-9052-49d4-bce8-9c16f0339382",
		Name:            "Advanced Hoverguard",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}: This creature gains shroud until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    effects.ManaCost("{U}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("shroud", "Advanced Hoverguard — shroud"),
		}},
	})
}
