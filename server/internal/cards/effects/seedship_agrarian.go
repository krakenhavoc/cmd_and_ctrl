package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "003d9c17-5d5c-469f-ad87-d78faf55e025",
		Name:         "Seedship Agrarian",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, ThisBecameTapped,
				"Seedship Agrarian — create a Lander token",
				Do(CreateToken{Template: b16LanderToken(), N: 1})),
			Landfall("Seedship Agrarian — put a +1/+1 counter on this creature (landfall)",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
