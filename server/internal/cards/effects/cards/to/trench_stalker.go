package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "5a0c215a-2164-4cb8-89ff-2558b6b42dca",
		Name:         "Trench Stalker",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.S82BroB07SelfKeywordWhileDrawnTwo("deathtouch"),
			effects.S82BroB07SelfKeywordWhileDrawnTwo("lifelink"),
		},
	})
}
