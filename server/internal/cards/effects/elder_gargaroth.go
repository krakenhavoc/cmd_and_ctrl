package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Elder Gargaroth — Creature — Beast {3}{G}{G}, 6/6:
//
//	"Reach, vigilance, trample
//	 Whenever this creature attacks or blocks, choose one —
//	 • Create a 3/3 green Beast creature token.
//	 • You gain 3 life.
//	 • Draw a card."
//
// One trigger on either event, with the mode chosen as it goes on the
// stack. The Beast arrives after attackers or blockers are declared, so
// it neither attacks nor blocks that combat.
//
// No simplification.
func init() {
	trigger := OnAny([]game.EventKind{game.EventAttack, game.EventBlock},
		func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return attackDeclared(ev, source) || selfBlocksOnce(ev, source)
		},
		"Elder Gargaroth — choose one", func(*game.Game, *game.StackItem) error { return nil })
	trigger.Modes = ChooseOne(
		ModeWithPurpose(ModeDoing("Create a 3/3 green Beast creature token.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return CreateToken{Controller: item.Controller, Template: TokenCard("3/3 green Beast"), N: 1}.Apply(ctx)
			}), game.Purpose{Tokens: 1, CreatureTokens: 1}),
		ModeWithPurpose(ModeDoing("You gain 3 life.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return GainLife{Player: item.Controller, Amount: 3}.Apply(ctx)
			}), game.Purpose{LifeGain: 3}),
		ModeWithPurpose(ModeDoing("Draw a card.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return ctx.Game.DrawNForEffect(item.Controller, 1)
			}), game.Purpose{Draws: 1}),
	)
	Register(Spec{
		OracleID:        "397b9dcf-690a-4a58-8637-bb9baab7cc2f",
		Name:            "Elder Gargaroth",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "vigilance", "trample"},
		Triggered:       []game.TriggeredAbility{trigger},
	})
}
