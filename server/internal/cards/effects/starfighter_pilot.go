package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Starfighter Pilot — Creature — Human Pilot {1}{W}, 2/2 (Edge of
// Eternities, #4461):
//
//	"Whenever this creature becomes tapped, surveil 1."
//
// Mechan Navigator's "becomes tapped" watch (ThisBecameTapped over a tap
// and an attack, so a vigilance attacker does not surveil) with a
// surveil 1. Tapping it for a cost (crew, station, convoke) surveils
// too.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fd6afb5e-d99a-44a7-9af4-9f064da95d63",
		Name:         "Starfighter Pilot",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, ThisBecameTapped,
				"Starfighter Pilot — surveil 1", Do(Surveil{N: 1})),
		},
	})
}
