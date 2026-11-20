package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "f5ef25fb-545e-41b7-8fed-d3cfff3d8668",
		Name:         "Ravenous Daggertooth",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			cmdr2B08EnrageTrigger("Ravenous Daggertooth — Enrage: you gain 2 life", Do(GainLife{Amount: 2})),
		},
	})
}
