package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brotherhood Patriarch — Creature — Human Assassin {3}{B}, 4/1
// (Assassin's Creed, #5466):
//
//	"When this creature dies, each opponent loses 2 life and you gain 2
//	 life."
//
// Symbiote Spawn's trigger: each opponent loses first, then you gain.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0c766a86-9c70-465e-874b-9d9f3b4eb090",
		Name:         "Brotherhood Patriarch",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Brotherhood Patriarch — each opponent loses 2 life and you gain 2 life",
				Do(EachOpponentLosesLife{N: 2}, GainLife{Amount: 2})),
		},
	})
}
