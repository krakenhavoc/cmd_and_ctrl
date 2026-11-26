package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Forlorn Pseudamma — {3}{B} Creature — Zombie 2/1 (Born of the Gods):
//
//	"Intimidate
//	 Inspired — Whenever this creature becomes untapped, you may pay {2}{B}. If you do,
//	 create a 2/2 black Zombie enchantment creature token."
//
// Aerie Worshippers' shape: the payment is asked as the trigger resolves.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "92680de3-6ee2-4b73-b0f5-38383482dcca",
		Name:            "Forlorn Pseudamma",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"intimidate"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventUntapCard, effects.Self, "Forlorn Pseudamma — you may pay {2}{B}; if you do, create a 2/2 black Zombie enchantment creature token (inspired)",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{2}{B}",
						Question: "Forlorn Pseudamma — pay {2}{B} to create a 2/2 black Zombie enchantment creature token?",
						Then:     forlornPseudammaPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// forlornPseudammaPaid is the paid branch, keyed: an on-disk identity, never renamed, never reused.
var forlornPseudammaPaid = effects.PayUnlessBranches("pay-unless/forlorn-pseudamma-zombie",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.CreateToken{Template: effects.TokenCard("2/2 black Zombie enchantment creature"), N: 1}.Apply(ctx)
	}, nil)
