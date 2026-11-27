package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Snake Pit — Enchantment {3}{G} (Mercadian Masques, S82 OLD batch 699, #5619):
//
//	"Whenever an opponent casts a blue or black spell, you may create a 1/1
//	 green Snake creature token."
//
// Mold Adder's cast trigger (AnOpponentCast over blue or black, a spell that is
// both counts once), with a real "may" asked as the trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "115bd33d-4875-4014-ae9f-0052f5acdbb2",
		Name:         "Snake Pit",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventCast, effects.AnOpponentCast(effects.Or(effects.OfColor("U"), effects.OfColor("B"))),
				"Snake Pit — create a 1/1 green Snake creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green Snake"), N: 1})),
				"Snake Pit — create a 1/1 green Snake creature token?"),
		},
	})
}
