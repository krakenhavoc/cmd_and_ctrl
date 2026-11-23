package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vision Skeins — Instant {1}{U} (EDHREC rank 4161):
//
//	"Each player draws two cards."
//
// A group-hug cantrip, and in a four-player game a genuinely
// symmetrical one: two mana buys you two cards and buys each of your
// three opponents two as well. It is played by decks that profit from
// other people drawing — Nekusar, Rhystic Study, Consecrated Sphinx,
// Notion Thief — where the symmetry is the point rather than the
// price.
//
// The draws go APNAP from the active seat and eliminated seats are
// skipped, which is b05EachPlayerDraws' contract; each one is an
// ordinary draw, so every "whenever an opponent draws" watcher on the
// table sees all six of them and a Notion Thief replaces them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6d42dff5-97ec-4768-b112-84a3584d0b87",
		Name:         "Vision Skeins",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.B05EachPlayerDraws(ctx.Game, item, 2)
		},
	})
}
