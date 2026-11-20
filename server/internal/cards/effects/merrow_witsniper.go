package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "79f8378a-fb50-4178-be53-b6926bc1cba5",
		Name:         "Merrow Witsniper",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Merrow Witsniper — target player mills a card", s82BNSB08MillTargetPlayer),
				TargetPlayer("target player")),
		},
	})
}
