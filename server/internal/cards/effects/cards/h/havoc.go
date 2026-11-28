package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Havoc — {1}{R} Enchantment (Tempest, S82 OLD batch 769, #5628):
//
//	"Whenever an opponent casts a white spell, they lose 2 life."
//
// One trigger per white spell an opponent casts; "they" is that opponent, who
// loses the life (a life loss the enchantment causes, not damage).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d95aeeb0-b722-4f70-85fc-6866be4854bd",
		Name:         "Havoc",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.AnOpponentCast(effects.OfColor("W")),
				"Havoc — that player loses 2 life", effects.S82OldB769OpponentLosesLife(2)),
		},
	})
}
