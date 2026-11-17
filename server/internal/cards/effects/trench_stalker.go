package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Trench Stalker — Creature — Beast Horror {4}{B}, 4/5 (The Brothers' War,
// #5482):
//
//	"As long as you've drawn two or more cards this turn, this creature
//	 has deathtouch and lifelink."
//
// Two layer 6 grants to itself read off the turn's draw tally (Foggy Swamp
// Hunters' count); the tally is not on the battlefield, so each declares
// DependsOnHandSize.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5a0c215a-2164-4cb8-89ff-2558b6b42dca",
		Name:         "Trench Stalker",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			s82BroB07SelfKeywordWhileDrawnTwo("deathtouch"),
			s82BroB07SelfKeywordWhileDrawnTwo("lifelink"),
		},
	})
}
