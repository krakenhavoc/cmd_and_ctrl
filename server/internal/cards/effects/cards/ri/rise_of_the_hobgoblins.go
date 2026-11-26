package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rise of the Hobgoblins — Enchantment {R/W}{R/W}:
//
//	"When this enchantment enters, you may pay {X}. If you do, create X 1/1
//	 red and white Goblin Soldier creature tokens.
//	 {R/W}: Red creatures and white creatures you control gain first strike
//	 until end of turn."
//
// Decree of Justice's pay-{X} prompt and token step: the payment is made as
// the trigger resolves and X may be zero. The activated ability gives first
// strike to the red creatures and the white creatures you control as it
// resolves (CR 611.2c); a creature that is both is one creature.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dfd3adec-033b-47ed-a227-ea1f48b4057c",
		Name:         "Rise of the Hobgoblins",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Rise of the Hobgoblins — you may pay {X}; if you do, create X 1/1 red and white Goblin Soldier creature tokens",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPayX{
						Cost:  "{X}",
						Label: "Rise of the Hobgoblins",
						Buys:  "create X 1/1 red and white Goblin Soldier creature tokens",
						Goal:  effects.XAsHighAsYouCan,
						Then:  riseOfTheHobgoblinsPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R/W}: Red creatures and white creatures you control gain first strike until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{R/W}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.GrantKeywordUntilEOT{
					Match:    effects.And(effects.Creature(), effects.YouControl(), effects.Or(effects.OfColor("R"), effects.OfColor("W"))),
					Keywords: []string{"first strike"},
					Label:    "Rise of the Hobgoblins — red and white creatures gain first strike until end of turn",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}

var riseOfTheHobgoblinsPaid = effects.PayUnlessBranches("pay-unless/rise-of-the-hobgoblins-goblins",
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		return effects.CreateToken{Template: effects.TokenCard("1/1 red and white Goblin Soldier"), N: r.Number}.Apply(ctx)
	}, nil)
