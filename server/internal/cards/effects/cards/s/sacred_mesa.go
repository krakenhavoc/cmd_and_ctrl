package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sacred Mesa — Enchantment {2}{W} (Commander 2014):
//
//	"At the beginning of your upkeep, sacrifice this enchantment unless you
//	 sacrifice a Pegasus.
//	 {1}{W}: Create a 1/1 white Pegasus creature token with flying."
//
// Contamination's tax over Pegasuses: with no Pegasus to give, the Mesa
// is sacrificed outright; otherwise the controller picks which to give up.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8776d14f-a5a2-4ef3-98d0-e1148e579e4d",
		Name:         "Sacred Mesa",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Sacred Mesa — sacrifice this enchantment unless you sacrifice a Pegasus",
				effects.SacrificeThisUnlessYouSacrificeA("Sacred Mesa", "a Pegasus", effects.Subtype("Pegasus"), effects.SacrificeAPegasusOrThisAnswer)),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}: Create a 1/1 white Pegasus creature token with flying.",
			Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerMakesBlocker},
			Cost:    effects.ManaCost("{1}{W}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CreateToken{Template: effects.TokenCard("1/1 white Pegasus with flying"), N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
