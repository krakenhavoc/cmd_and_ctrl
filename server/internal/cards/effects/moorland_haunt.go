package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Moorland Haunt — Land (EDHREC rank 8707):
//
//	"{T}: Add {C}.
//	 {W}{U}, {T}, Exile a creature card from your graveyard: Create a
//	 1/1 white Spirit creature token with flying."
//
// The Azorius creature-land-that-isn't. The second ability's cost is
// AbilityCost.ExileCards with a predicate (#1297): only a creature card
// in your own graveyard can pay, and the picker, the enumerator and the
// validator all read the one candidate walk, so a graveyard holding
// only spells offers nothing to pay with and the ability is not
// offered. The Spirit is the same flying 1/1 Bishop of Wings makes.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5324192b-6687-41e4-8e56-326b21a5dbf3",
		Name:         "Moorland Haunt",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label: "{W}{U}, {T}, Exile a creature card from your graveyard: Create a 1/1 white Spirit creature token with flying.",
			Cost:  Plus(ManaCost("{W}{U}"), TapCost(), ExileFromGraveyard(1, "a creature card", MatchCreature)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return CreateToken{Controller: item.Controller, Template: b28WhiteSpiritFlyingToken(), N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
