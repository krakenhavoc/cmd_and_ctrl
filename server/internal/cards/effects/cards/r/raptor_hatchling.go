package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Raptor Hatchling — Creature — Dinosaur {1}{R}, 1/1 (Ixalan, #5532):
//
//	"Enrage — Whenever this creature is dealt damage, create a 3/3 green Dinosaur creature token with trample."
//
// Enrage is the shared trigger (cmdr2B08EnrageTrigger, Bellowing
// Aegisaur's): it triggers once for each event that deals this creature
// damage, however much, and still resolves if the damage was lethal.
// Its controller creates a 3/3 green
// Dinosaur with trample.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "91638137-06ba-4b21-b088-28d2b4d1d8da",
		Name:         "Raptor Hatchling",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.Cmdr2B08EnrageTrigger("Raptor Hatchling — Enrage: create a 3/3 green Dinosaur with trample",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("3/3 green Dinosaur with trample"), N: 1})),
		},
	})
}
