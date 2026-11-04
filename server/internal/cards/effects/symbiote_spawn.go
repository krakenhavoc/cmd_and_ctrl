package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Symbiote Spawn — Creature — Symbiote Villain {2}{B}, 3/2 (Marvel's
// Spider-Man, #4460):
//
//	"When this creature dies, each opponent loses 2 life and you gain 2
//	 life."
//
// "You" is the controller as it last existed (CR 603.10a), which the
// trigger item carries. Each opponent loses first, then you gain.
func init() {
	Register(Spec{
		OracleID:     "7c3ff360-461a-44f3-a8f2-eb061b149cf5",
		Name:         "Symbiote Spawn",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Symbiote Spawn — each opponent loses 2 life and you gain 2 life",
				Do(EachOpponentLosesLife{N: 2}, GainLife{Amount: 2})),
		},
	})
}
