package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chambered Nautilus — Creature — Nautilus Beast {2}{U}, 2/2 (Mercadian
// Masques, S82 OLD, #5619):
//
//	"Whenever this creature becomes blocked, you may draw a card."
//
// "Becomes blocked" is once per blocked attacker (CR 509.3c). The "may" is
// asked as the trigger resolves, and a player a draw limit stops drawing is
// not asked (#4016).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0e1267d4-0155-4503-829a-03264c8d14a0",
		Name:         "Chambered Nautilus",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OptionalDraw(effects.On(game.EventBecomesBlocked, effects.Self,
				"Chambered Nautilus — you may draw a card", effects.Do(effects.DrawCards{N: 1})),
				"Chambered Nautilus — draw a card?", 1),
		},
	})
}
