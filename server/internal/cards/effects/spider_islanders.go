package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spider-Islanders — Creature — Spider Horror Citizen {3}{R}, 4/3 (Marvel's
// Spider-Man, #4460):
//
//	"Mayhem {1}{R} (You may cast this card from your graveyard for {1}{R} if
//	 you discarded it this turn. Timing rules still apply.)"
//
// Swarm, Being of Bees' shape: the engine's mayhem keyword is the graveyard cast
// zone while discarded this turn, plus the alternative cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "2d1a15dc-3123-4425-b036-a3b5a83fcdf8",
		Name:             "Spider-Islanders",
		Completeness:     CompletenessFull,
		CastableZonesIf:  []game.ConditionalCastZone{MayhemCastableZone()},
		AlternativeCosts: []game.AlternativeCost{Mayhem("{1}{R}")},
	})
}
