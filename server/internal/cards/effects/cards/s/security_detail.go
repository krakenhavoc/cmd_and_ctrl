package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Security Detail — Enchantment {3}{W} (Mercadian Masques, S82 OLD b681,
// #5619):
//
//	"{W}{W}: Create a 1/1 white Soldier creature token. Activate only if you
//	 control no creatures and only once each turn."
//
// Defenders of Humanity's gate (you control no creatures, checked at
// activation, CR 602.5b) and Astute Terraformer's once-each-turn tally.
//
// No simplification.
func init() {
	const label = "{W}{W}: Create a 1/1 white Soldier creature token. Activate only if you control no creatures and only once each turn."
	effects.Register(effects.Spec{
		OracleID:     "2a26dee4-b4b6-489d-99e5-6dbd8d19419d",
		Name:         "Security Detail",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   label,
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    effects.ManaCost("{W}{W}"),
			Condition: effects.AllConditions(effects.OncePerTurnActivation(label), func(g *game.Game, controller, _ uuid.UUID) bool {
				return effects.CountControlled(g, controller, func(c game.Card) bool { return c.IsCreature() }) == 0
			}),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.CreateToken{Controller: ctx.Controller(), Template: effects.TokenCard("1/1 white Soldier"), N: 1}.Apply(ctx)
			},
		}},
	})
}
