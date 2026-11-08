package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hidden Grotto — Land (Bloomburrow, #4467):
//
//	"When this land enters, surveil 1. (Look at the top card of your
//	 library. You may put it into your graveyard.)
//	 {T}: Add {C}.
//	 {1}, {T}: Add one mana of any color."
//
// Conduit Pylons without the Desert type: an untapped land whose surveil
// trigger goes on the stack, a colourless tap, and the {1}, {T} filter
// into the five-colour pipe.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "382d18a2-438e-4ae7-a83f-1658ef1f9b07",
		Name:         "Hidden Grotto",
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
			WhenThisEnters("Hidden Grotto — surveil 1", Do(Surveil{N: 1})),
		},
	})
}
