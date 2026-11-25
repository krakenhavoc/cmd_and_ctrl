package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Drowned Secrets — Enchantment {1}{U} (Guilds of Ravnica, #5525):
//
//	"Whenever you cast a blue spell, target player mills two cards."
//
// Staff of the Mind Magus' blue-spell cast trigger, with a player target chosen
// as the trigger goes on the stack (CR 603.3d): the pick, if still legal as it
// resolves, mills two cards. A spell cast by anyone else, and a spell that is
// not blue, does not trigger it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cab305d9-27fb-4a41-adf1-892e15167cb9",
		Name:         "Drowned Secrets",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverYouCast(effects.OfColor("U"), "Drowned Secrets — target player mills two cards",
				effects.TargetPlayerMillsEffect(2)),
				effects.TargetPlayer("target player")),
		},
	})
}
