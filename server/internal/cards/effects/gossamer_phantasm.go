package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gossamer Phantasm — Creature — Illusion {1}{U}, 2/1 (Planar Chaos, #5595):
//
//	"Flying
//	 When this creature becomes the target of a spell or ability, sacrifice
//	 it."
//
// Phantasmal Image's Illusion trigger, printed: it goes on the stack when
// the creature becomes a target (CR 603.2), so the spell or ability that
// targeted it is still on the stack beneath it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5ed23ed2-be2f-4958-affe-429f58e9dca5",
		Name:            "Gossamer Phantasm",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventBecomesTarget, Self, "Gossamer Phantasm — sacrifice it", SacrificeThisIfStillOnBattlefield),
		},
	})
}
