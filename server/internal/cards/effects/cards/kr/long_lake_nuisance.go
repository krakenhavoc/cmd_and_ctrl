package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Long Lake Nuisance — Creature — Bird {3}{U}, 3/1:
//
//	"Flying
//	 When this creature enters, recruit. (Draw a card, then discard a
//	 card. If you discarded a nonland card, create a 1/1 white Human
//	 Soldier creature token.)"
//
// No simplification. The Hobbit (#4475).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a833fdf1-db0c-4846-8452-d3b2059c2355",
		Name:            "Long Lake Nuisance",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Long Lake Nuisance — recruit", effects.Do(effects.Recruit{})),
		},
	})
}
