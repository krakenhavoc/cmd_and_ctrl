package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Warmth — {1}{W} Enchantment (Tempest, S82 OLD batch 769, #5628):
//
//	"Whenever an opponent casts a red spell, you gain 2 life."
//
// One trigger per red spell an opponent casts (Aether Sting's cast condition,
// read through the spell's colour on the stack); the life is its controller's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cccc9664-0e3d-44c4-959c-abeaba25bbab",
		Name:         "Warmth",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.AnOpponentCast(effects.OfColor("R")),
				"Warmth — you gain 2 life", effects.Do(effects.GainLife{Amount: 2})),
		},
	})
}
