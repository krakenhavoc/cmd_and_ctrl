package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ravenous Daggertooth — Creature — Dinosaur {2}{G}, 3/2 (Ixalan, #5532):
//
//	"Enrage — Whenever this creature is dealt damage, you gain 2 life."
//
// Enrage is the shared trigger (cmdr2B08EnrageTrigger, Bellowing
// Aegisaur's): it triggers once for each event that deals this creature
// damage, however much, and still resolves if the damage was lethal.
// Its controller gains 2 life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f5ef25fb-545e-41b7-8fed-d3cfff3d8668",
		Name:         "Ravenous Daggertooth",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Cmdr2B08EnrageTrigger("Ravenous Daggertooth — Enrage: you gain 2 life", effects.Do(effects.GainLife{Amount: 2})),
		},
	})
}
