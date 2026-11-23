package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Merrow Witsniper — Creature — Merfolk Rogue {U}, 1/1 (Morningtide,
// #5592):
//
//	"When this creature enters, target player mills a card."
//
// An enters trigger targeting any player, Duskmantle's "target player
// mills a card" body (s82BNSB08MillTargetPlayer); a player who left in
// response mills nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "79f8378a-fb50-4178-be53-b6926bc1cba5",
		Name:         "Merrow Witsniper",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Merrow Witsniper — target player mills a card", effects.S82BNSB08MillTargetPlayer),
				effects.TargetPlayer("target player")),
		},
	})
}
