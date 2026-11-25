package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Emrakul's Evangel — Creature — Human Horror {2}{G}, 3/2 (Eldritch Moon,
// S82 OLD, #5541):
//
//	"{T}, Sacrifice this creature and any number of other non-Eldrazi
//	 creatures: Create a 3/2 colorless Eldrazi Horror creature token for each
//	 creature sacrificed this way."
//
// Sword of the Ages' variable sacrifice (#6121): this creature as a cost
// beside "any number of" other creatures, none of which may be an Eldrazi.
// The count is read off the announcement (ctx.Sacrificed), this creature
// included, so a lone sacrifice of the Evangel makes one token (CR 602.2b).
//
// No simplification.
func init() {
	others := effects.Another(effects.SacrificeSpec("any number of other non-Eldrazi creatures",
		effects.Creature(), effects.Not(effects.HasSubtype("Eldrazi")))).WithCount(0, 0)
	effects.Register(effects.Spec{
		OracleID:     "99eb572b-85b1-4d78-aea8-4400a5b7e508",
		Name:         "Emrakul's Evangel",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice this creature and any number of other non-Eldrazi creatures: Create a 3/2 colorless Eldrazi Horror creature token for each creature sacrificed this way.",
			Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerSacOutlet | game.AnswerMakesBlocker},
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeThis(), game.AbilityCost{SacrificeOther: others}),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.CreateToken{Template: effects.TokenCard("3/2 colorless Eldrazi Horror"), N: max(ctx.Sacrificed(), 1)}.Apply(ctx)
			},
		}},
	})
}
