package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	sevenDevotion := CondBoard{
		Name: "your devotion to white and black is seven or greater",
		Test: func(ctx *Context) bool {
			return devotionToColors(ctx.Game, ctx.Controller(), "W", "B") >= 7
		},
	}
	Register(Spec{
		OracleID:     "9eb427ab-1483-4c40-8cb2-fd463e1e5fe0",
		Name:         "March of the Canonized",
		Completeness: CompletenessFull,
		XMatters:     true,
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: Self,
				Key:       "March of the Canonized — create X 1/1 white Vampire creature tokens with lifelink",
				XFromCast: true,
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return CreateToken{Template: TokenCard("1/1 white Vampire with lifelink"), N: ctx.X()}.Apply(ctx)
				},
			},
			InterveningIf(sevenDevotion, On(game.EventBeginUpkeep,
				func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
					return ByYou(ev, source, ch, g) && devotionToColors(g, source.Controller, "W", "B") >= 7
				},
				"March of the Canonized — create a 4/3 white and black Vampire Demon creature token with flying",
				Do(CreateToken{Template: TokenCard("4/3 white and black Vampire Demon with flying"), N: 1}))),
		},
	})
}
