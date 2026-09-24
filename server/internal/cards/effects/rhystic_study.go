package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rhystic Study — Enchantment for {2}{U}:
//
//	"Whenever an opponent casts a spell, you may draw a card unless
//	that player pays {1}."
//
// S19 sub-PR 6: the first cast trigger on the auto-fire pipeline,
// and the first pay-unless card. EventCast carries the caster in
// Actor; AppliesTo gates on Actor != Controller. The trigger goes
// on the stack; on resolution it asks the caster "pay {1}?" via
// PayUnless. Declining (or answering yes without the mana) asks the
// Study's controller a second, independent question — "draw a
// card?" — via MayChoice (#796), which is exactly the "you may [do
// X] mid-resolution" primitive this needed: the decision comes
// after the caster's answer, addressed to a different seat, neither
// a search nor a cost the engine already prompts for. A controller
// with an empty library can decline and stay alive; declining a
// "you may draw" that would kill you is a real line in paper and
// now a real choice here.
func init() {
	Register(Spec{
		OracleID:     "53236dd7-845a-444c-96d5-f41ed7325d8f",
		Name:         "Rhystic Study",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor != source.Controller
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				caster := ev.Actor
				return game.NewTriggeredItem(source, "Rhystic Study — draw unless caster pays {1}",
					func(g *game.Game, item *game.StackItem) error {
						return PayUnless{
							Chooser:  caster,
							Cost:     "{1}",
							Question: "Rhystic Study — pay {1}?",
							OnDecline: func(ctx *Context) error {
								return MayChoice{
									Question: "Rhystic Study — draw a card?",
									OnYes: func(ctx *Context) error {
										return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
									},
								}.Apply(ctx)
							},
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
