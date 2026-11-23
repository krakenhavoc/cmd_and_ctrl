package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Daring Skyjek — Creature — Human Knight {1}{W}, 3/1 (Gatecrash, #5565):
//
//	"Battalion — Whenever this creature and at least two other creatures
//	 attack, this creature gains flying until end of turn."
//
// Battalion is Boros Strike-Captain's group test (gtcBattalion); the
// bonus is pinned to this creature and ends in the cleanup step.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f15a202d-c1c4-49bd-924a-7a0a5667ff70",
		Name:         "Daring Skyjek",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.GtcBattalion("Daring Skyjek — it gains flying until end of turn (battalion)",
				effects.ThisGainsKeywordUntilEndOfTurn("flying", "Daring Skyjek — flying until end of turn")),
		},
	})
}
