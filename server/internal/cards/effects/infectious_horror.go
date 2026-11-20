package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Infectious Horror — Creature — Zombie Horror {3}{B}, 2/2 (Conflux,
// #5587):
//
//	"Whenever this creature attacks, each opponent loses 2 life."
//
// Pulse Tracker's attack trigger at 2. It is life loss, not damage, so
// nothing prevents it, and every opponent loses it whichever player the
// creature attacks.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0931206b-eed2-40d0-9496-5ecefc7f8f90",
		Name:         "Infectious Horror",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, ThisAttacked,
				"Infectious Horror — each opponent loses 2 life",
				Do(EachOpponentLosesLife{N: 2})),
		},
	})
}
