package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Surveillance Room — Land (Marvel Super Heroes, #4455):
//
//	"When this land enters, surveil 1.
//	 {T}: Add {C}.
//	 {1}, {T}: Add one mana of any color."
//
// Conduit Pylons' shape: the colorless mana ability, the filter ability
// and a surveil trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a3648376-dc8b-409b-b2d1-c29e326a059c",
		Name:         "Surveillance Room",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			painlessColorless(),
			{
				Cost:     ManaAbilityCost{Tap: true, Mana: "{1}"},
				Produced: "{W|U|B|R|G}",
				Label:    "{1}, {T}: Add one mana of any color",
			},
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Surveillance Room — surveil 1", Do(Surveil{N: 1})),
		},
	})
}
