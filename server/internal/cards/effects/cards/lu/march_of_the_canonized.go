package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// March of the Canonized — Enchantment {X}{W}{W} (Commander, #5368):
//
//	"When this enchantment enters, create X 1/1 white Vampire creature
//	 tokens with lifelink.
//	 At the beginning of your upkeep, if your devotion to white and black
//	 is seven or greater, create a 4/3 white and black Vampire Demon
//	 creature token with flying."
//
// X is the cast's X, stamped on the entry trigger (XFromCast, as
// Champions from Beyond's is). The upkeep trigger is an intervening "if"
// (CR 603.4): devotion to white and black counts every {W} and {B} mana
// symbol (a hybrid symbol of either colour once) among the mana costs of
// permanents you control, as it triggers and again as it resolves.
//
// No simplification.
func init() {
	sevenDevotion := effects.CondBoard{
		Name: "your devotion to white and black is seven or greater",
		Test: func(ctx *effects.Context) bool {
			return effects.DevotionToColors(ctx.Game, ctx.Controller(), "W", "B") >= 7
		},
	}
	effects.Register(effects.Spec{
		OracleID:     "9eb427ab-1483-4c40-8cb2-fd463e1e5fe0",
		Name:         "March of the Canonized",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: effects.Self,
				Key:       "March of the Canonized — create X 1/1 white Vampire creature tokens with lifelink",
				XFromCast: true,
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.CreateToken{Template: effects.TokenCard("1/1 white Vampire with lifelink"), N: ctx.X()}.Apply(ctx)
				},
			},
			effects.InterveningIf(sevenDevotion, effects.On(game.EventBeginUpkeep,
				func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
					return effects.ByYou(ev, source, ch, g) && effects.DevotionToColors(g, source.Controller, "W", "B") >= 7
				},
				"March of the Canonized — create a 4/3 white and black Vampire Demon creature token with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("4/3 white and black Vampire Demon with flying"), N: 1}))),
		},
	})
}
