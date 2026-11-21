package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dawn of Hope — Enchantment {1}{W} (EDHREC rank 2417):
//
//	"Whenever you gain life, you may pay {2}. If you do, draw a card.
//	 {3}{W}: Create a 1/1 white Soldier creature token with lifelink."
//
// The lifegain deck's card-draw engine, feeding itself. The trigger
// is Marauding Blight-Priest's condition (every lifegain path emits
// EventChangeLife with a positive Amount, lifelink included) with
// the MayPay primitive — a prompt for the controller that draws on
// "Pay" and does nothing on "Don't pay". Once per life-gain event,
// so three lifelink creatures connecting are three prompts, as
// printed. The activation is an ordinary mana-cost ability making a
// Soldier whose lifelink feeds the trigger next combat.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d7a38484-2acc-49a6-b32d-d54dddb14d31",
		Name:         "Dawn of Hope",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouGainLife("Dawn of Hope — pay {2} to draw a card", func(g *game.Game, item *game.StackItem) error {
				return effects.MayPay{
					Chooser:  item.Controller,
					Cost:     "{2}",
					Question: "Dawn of Hope — pay {2} to draw a card?",
					Then:     effects.MayPayDrawACard,
				}.Apply(effects.NewContext(g, item))
			}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{W}: Create a 1/1 white Soldier creature token with lifelink.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    effects.ManaCost("{3}{W}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("1/1 white Soldier with lifelink"), N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
