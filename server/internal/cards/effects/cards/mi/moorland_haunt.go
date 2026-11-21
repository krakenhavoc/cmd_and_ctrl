package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "5324192b-6687-41e4-8e56-326b21a5dbf3",
		Name:         "Moorland Haunt",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}{U}, {T}, Exile a creature card from your graveyard: Create a 1/1 white Spirit creature token with flying.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    effects.Plus(effects.ManaCost("{W}{U}"), effects.TapCost(), effects.ExileFromGraveyard(1, "a creature card", effects.MatchCreature)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CreateToken{Controller: item.Controller, Template: effects.B28WhiteSpiritFlyingToken(), N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
