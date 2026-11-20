package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Phantasmal Dragon — Creature — Dragon Illusion {2}{U}{U}, 5/5 (Magic 2012,
// #5573):
//
//	"Flying
//	 When this creature becomes the target of a spell or ability,
//	 sacrifice it."
//
// Flying, and Phantasmal Shieldback's Illusion clause: a trigger on the
// Dragon being targeted by any spell or ability, whose resolution
// sacrifices it if it is still the same permanent.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7f32c581-4d4a-4873-8e4b-b95e13790481",
		Name:            "Phantasmal Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventBecomesTarget, Self,
				"Phantasmal Dragon — sacrifice it",
				SacrificeThisIfStillOnBattlefield),
		},
	})
}
