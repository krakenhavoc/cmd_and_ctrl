package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bonepicker Skirge — Creature — Phyrexian Imp {2}{B}, 2/2:
//
//	"Flying
//	 Corrupted — As long as an opponent has three or more poison counters,
//	 this creature has deathtouch and lifelink."
//
// One printed sentence, two layer 6 grants on the same live condition.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c42833e6-3e42-4bd6-acbc-d8e284e9833e",
		Name:            "Bonepicker Skirge",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			effects.SelfKeywordWhileCorrupted("deathtouch"),
			effects.SelfKeywordWhileCorrupted("lifelink"),
		},
	})
}
