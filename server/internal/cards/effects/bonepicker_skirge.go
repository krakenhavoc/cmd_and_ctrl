package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "c42833e6-3e42-4bd6-acbc-d8e284e9833e",
		Name:            "Bonepicker Skirge",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			SelfKeywordWhileCorrupted("deathtouch"),
			SelfKeywordWhileCorrupted("lifelink"),
		},
	})
}
