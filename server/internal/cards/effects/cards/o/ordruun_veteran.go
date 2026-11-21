package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ordruun Veteran — Creature — Minotaur Soldier {2}{R}{W}, 3/1 (Gatecrash, #5565):
//
//	"Battalion — Whenever this creature and at least two other creatures
//	 attack, this creature gains double strike until end of turn. (It deals
//	 both first-strike and regular combat damage.)"
//
// Battalion is Boros Strike-Captain's group test (gtcBattalion); the
// bonus is pinned to this creature and ends in the cleanup step.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "119af12a-b767-4408-9bb5-72a6d0ec63d1",
		Name:         "Ordruun Veteran",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.GtcBattalion("Ordruun Veteran — it gains double strike until end of turn (battalion)",
				effects.ThisGainsKeywordUntilEndOfTurn("double strike", "Ordruun Veteran — double strike until end of turn")),
		},
	})
}
