package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spawning Bed — Land (Modern Horizons 3 Commander, #5948):
//
//	"{T}: Add {C}.
//	 {6}, {T}, Sacrifice this land: Create three 1/1 colorless Eldrazi Scion
//	 creature tokens. They have "Sacrifice this token: Add {C}.""
//
// The Scions are the catalog's Eldrazi Scion (b28EldraziScionToken), whose
// "Sacrifice this token: Add {C}" is a real mana ability. The cost is paid
// whole before anything resolves, so the Scions arrive even if the land was
// the last thing sacrificed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "49e43de3-460b-4562-aef6-da43bd56debc",
		Name:         "Spawning Bed",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{6}, {T}, Sacrifice this land: Create three 1/1 colorless Eldrazi Scion creature tokens. They have \"Sacrifice this token: Add {C}.\"",
			Purpose: game.Purpose{Tokens: 3, CreatureTokens: 3, Answers: game.AnswerMakesBlocker},
			Cost:    effects.Plus(effects.ManaCost("{6}"), effects.TapCost(), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CreateToken{Controller: item.Controller, Template: effects.B28EldraziScionToken(), N: 3}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
