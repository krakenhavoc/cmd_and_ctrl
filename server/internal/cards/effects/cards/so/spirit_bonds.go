package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spirit Bonds — Enchantment {1}{W} (Magic 2015, #5556):
//
//	"Whenever a nontoken creature you control enters, you may pay {W}. If you
//	 do, create a 1/1 white Spirit creature token with flying.
//	 {1}{W}, Sacrifice a Spirit: Target non-Spirit creature gains
//	 indestructible until end of turn."
//
// Ruthless Sniper's may-pay trigger over any nontoken creature that enters
// under your control (itself included, and a token never counts), and a
// sacrifice-a-Spirit cost: any Spirit you control, token or not, read through
// the layers, and the target can't be a Spirit.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f41ea1b5-2c31-43fa-b142-244add1efd35",
		Name:         "Spirit Bonds",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
				return ok && c.IsCreature() && !c.IsToken()
			}, "Spirit Bonds — you may pay {W}; if you do, create a 1/1 white Spirit creature token with flying",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{W}",
						Question: "Spirit Bonds — pay {W} to create a 1/1 white Spirit creature token with flying?",
						Then:     spiritBondsPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}, Sacrifice a Spirit: Target non-Spirit creature gains indestructible until end of turn.",
			Cost:    effects.Plus(effects.ManaCost("{1}{W}"), game.AbilityCost{SacrificeOther: effects.SacrificeSpec("a Spirit", effects.Subtype("Spirit"))}),
			Targets: effects.TargetCreature("target non-Spirit creature", effects.Not(effects.Subtype("Spirit"))),
			Purpose: effects.WithAnswers(effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "indestructible")), game.AnswerProtect),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				targets := ctx.LegalTargets()
				if len(targets) == 0 {
					return nil
				}
				return effects.GrantKeywordUntilEOT{Target: targets[0].ID, Keywords: []string{"indestructible"}, Label: "Spirit Bonds — indestructible until end of turn"}.Apply(ctx)
			},
		}},
	})
}

// spiritBondsPaid creates the Spirit once {W} was really paid. An on-disk key:
// never renamed, never reused.
var spiritBondsPaid = effects.PayUnlessBranches("pay-unless/spirit-bonds-spirit", func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
	return effects.CreateToken{Template: effects.TokenCard("1/1 white Spirit with flying"), N: 1}.Apply(ctx)
}, nil)
