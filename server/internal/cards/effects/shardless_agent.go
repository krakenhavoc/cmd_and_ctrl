package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shardless Agent — Artifact Creature — Human Rogue {1}{G}{U}, 2/2:
//
//	"Cascade"
//
// Three mana, so it cascades into two or less — the tightest window
// of the group, and the reason a Shardless deck is built almost
// entirely out of one- and two-drops.
func init() {
	Register(Spec{
		OracleID:  "2afbaa9a-c171-4a8b-90f3-5250d8498356",
		Name:      "Shardless Agent",
		Triggered: []game.TriggeredAbility{Cascade()},
	})
}
