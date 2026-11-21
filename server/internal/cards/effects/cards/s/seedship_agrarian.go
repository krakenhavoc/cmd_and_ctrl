package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Seedship Agrarian — Creature — Insect Scientist {3}{G}, 3/3 (Edge of
// Eternities, #4461):
//
//	"Whenever this creature becomes tapped, create a Lander token.
//	 Landfall — Whenever a land you control enters, put a +1/+1 counter
//	 on this creature."
//
// Starfighter Pilot's "becomes tapped" watch (ThisBecameTapped over a tap
// and an attack, so a vigilance attacker makes nothing) with Sami's
// Curiosity's Lander token. Tapping it for a cost (crew, station,
// convoke) makes one too. The landfall counter goes on this creature
// only while it is still the same object (CounterOnThis).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "003d9c17-5d5c-469f-ad87-d78faf55e025",
		Name:         "Seedship Agrarian",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, effects.ThisBecameTapped,
				"Seedship Agrarian — create a Lander token",
				effects.Do(effects.CreateToken{Template: effects.B16LanderToken(), N: 1})),
			effects.Landfall("Seedship Agrarian — put a +1/+1 counter on this creature (landfall)",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
