package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hammer of Purphoros — Legendary Enchantment Artifact {1}{R}{R}
// (EDHREC rank 3074):
//
//	"Creatures you control have haste.
//	 {2}{R}, {T}, Sacrifice a land: Create a 3/3 colorless Golem
//	 enchantment artifact creature token."
//
// The haste anthem with a late-game land sink. The static is
// Urabrask's shape (b16GrantKeywords over b16CreaturesYouControl —
// layer 6, so the engine's summoning-sickness checks read it); the
// activation is a CR 602 ability with mana, tap and a
// sacrifice-a-land cost (b29SacrificeALand, the activator's own
// lands, post-layer types), and the Golem is its own template — an
// enchantment artifact creature, so a constellation trigger and an
// artifact-ETB payoff both see it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "212d058b-69c6-4dc2-8c93-bdfe26dc2ffe",
		Name:         "Hammer of Purphoros",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.B16GrantKeywords(effects.B16CreaturesYouControl, "haste"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}, {T}, Sacrifice a land: Create a 3/3 colorless Golem enchantment artifact creature token.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    effects.Plus(effects.ManaCost("{2}{R}"), effects.TapCost(), game.AbilityCost{SacrificeOther: effects.B29SacrificeALand()}),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("3/3 colorless Golem artifact"), N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
