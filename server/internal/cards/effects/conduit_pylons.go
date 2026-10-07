package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Conduit Pylons — Land — Desert:
//
//	"When this land enters, surveil 1. (Look at the top card of your
//	 library. You may put it into your graveyard.)
//	 {T}: Add {C}.
//	 {1}, {T}: Add one mana of any color."
//
// The surveil is the same ETB trigger the surveil lands use (it goes on
// the stack, so opponents can respond), and the filter ability is the
// Signet-style `ManaAbilityCost.Mana` component on a land, producing the
// five-colour pipe so the controller gets the colour picker. The land
// enters untapped, so no replacement is declared. Its Desert subtype
// derives no mana: the synthetic basic-land ability only reads basic
// land types.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "37f924e1-7c25-4f06-88bb-054693a21e5a",
		Name:         "Conduit Pylons",
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
			WhenThisEnters("Conduit Pylons — surveil 1", Do(Surveil{N: 1})),
		},
	})
}
