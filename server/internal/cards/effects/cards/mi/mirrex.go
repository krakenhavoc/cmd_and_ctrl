package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mirrex — Land — Sphere (Phyrexia: All Will Be One, #5479):
//
//	"{T}: Add {C}.
//	 {T}: Add one mana of any color. Activate only if this land entered
//	 this turn.
//	 {3}, {T}: Create a 1/1 colorless Phyrexian Mite artifact creature
//	 token with toxic 1 and 'This token can't block.'"
//
// The second mana ability carries the "entered this turn" clause as its
// Condition (game.EnteredThisTurn, not summoning sickness: CR 302.6
// would keep it on for a land that arrived on an opponent's turn). The
// token is the catalog's printed Phyrexian Mite.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5502741a-e3b9-454e-8121-4360a6db6750",
		Name:         "Mirrex",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			{Cost: effects.ManaAbilityCost{Tap: true}, Produced: "{C}", Label: "Add {C}"},
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{W|U|B|R|G}",
				Label:    "{T}: Add one mana of any color. Activate only if this land entered this turn.",
				Condition: func(g *game.Game, _, source uuid.UUID) bool {
					return g.EnteredThisTurn(source)
				},
			},
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: Create a 1/1 colorless Phyrexian Mite artifact creature token with toxic 1 and \"This token can't block.\"",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CreateToken{Template: effects.PhyrexianMiteToken(), N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
