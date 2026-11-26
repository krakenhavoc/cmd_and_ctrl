package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aerie Worshippers — {3}{U} Creature — Human Cleric 2/4 (Born of the Gods):
//
//	"Inspired — Whenever this creature becomes untapped, you may pay {2}{U}. If you do,
//	 create a 2/2 blue Bird enchantment creature token with flying."
//
// Inspired is an ability word: the trigger is Sphinx's Disciple's (EventUntapCard on
// itself), firing in the untap step and on any other untap. The payment is asked as
// the trigger resolves.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "674c34eb-1080-4811-8f2a-5006e08d98b8",
		Name:         "Aerie Worshippers",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventUntapCard, effects.Self, "Aerie Worshippers — you may pay {2}{U}; if you do, create a 2/2 blue Bird enchantment creature token with flying (inspired)",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{2}{U}",
						Question: "Aerie Worshippers — pay {2}{U} to create a 2/2 blue Bird enchantment creature token with flying?",
						Then:     aerieWorshippersPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// aerieWorshippersPaid is the paid branch, keyed: an on-disk identity, never renamed, never reused.
var aerieWorshippersPaid = effects.PayUnlessBranches("pay-unless/aerie-worshippers-bird",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.CreateToken{Template: effects.TokenCard("2/2 blue Bird enchantment with flying"), N: 1}.Apply(ctx)
	}, nil)
