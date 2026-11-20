package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Frost Walker — Creature — Elemental {1}{U}, 4/1 (Fate Reforged,
// #5552):
//
//	"When this creature becomes the target of a spell or ability,
//	 sacrifice it."
//
// Cephalid Illusionist's watch with Makeshift Mannequin's body: the trigger
// fires when it becomes the target of any spell or ability, and the
// sacrifice does nothing if it has already left the battlefield (CR
// 701.21a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "54a22814-c0d5-4637-9aa9-a6bc96401914",
		Name:         "Frost Walker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBecomesTarget, Self, "Frost Walker — sacrifice it", SacrificeThisIfStillOnBattlefield),
		},
	})
}
