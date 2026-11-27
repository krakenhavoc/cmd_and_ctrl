package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Deepwood Wolverine — Creature — Wolverine {G}, 1/1 (Mercadian Masques, S82
// OLD, #5619):
//
//	"Whenever this creature becomes blocked, it gets +2/+0 until end of turn."
//
// Snorting Gahr's trigger for +2/+0: once per blocked attacker (CR 509.3c), the
// bonus the Wolverine's own while it is still the same object (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ca8060a7-30e7-49a4-a720-da5b995dd31f",
		Name:         "Deepwood Wolverine",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesBlocked, effects.Self, "Deepwood Wolverine — it gets +2/+0 until end of turn",
				effects.ThisGetsUntilEndOfTurn(2, 0, "Deepwood Wolverine — +2/+0 until end of turn")),
		},
	})
}
