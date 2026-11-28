package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Council of Advisors — Creature — Human Advisor {2}{U}, 1/1 (Portal Three Kingdoms, S82 OLD b714, #5622):
//
//	"When this creature enters, draw a card."
//
// An enters trigger on the stack (Advancing the Spirit's shape).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ff427657-173a-4845-89af-6f0a93467130",
		Name:         "Council of Advisors",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Council of Advisors — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
