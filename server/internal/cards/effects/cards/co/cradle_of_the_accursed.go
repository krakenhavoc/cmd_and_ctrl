package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cradle of the Accursed — Land — Desert (Amonkhet, #5535):
//
//	"{T}: Add {C}.
//	 {3}, {T}, Sacrifice this land: Create a 2/2 black Zombie creature token.
//	 Activate only as a sorcery."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "36d06c91-5080-4f97-8e4c-ca8ac390e808",
		Name:          "Cradle of the Accursed",
		Completeness:  effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{effects.ColorlessTapRow()},
		Activated: []effects.ActivatedAbility{{
			Label:        "{3}, {T}, Sacrifice this land: Create a 2/2 black Zombie creature token. Activate only as a sorcery.",
			Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
			Cost:         effects.Plus(effects.ManaCost("{3}"), effects.TapCost(), effects.SacrificeThis()),
			SorcerySpeed: true,
			Effect:       effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 black Zombie"), N: 1}),
		}},
	})
}
