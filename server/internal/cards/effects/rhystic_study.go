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
// PayUnless. Declining (or answering yes without the mana) draws
// the Study's controller a card.
//
// Sandbox simplification: the controller's "you may draw" is
// treated as "draw" — a Rhystic player who doesn't want the card
// is rare enough that the extra prompt isn't worth the click. If
// that ever matters (empty library), DrawCards no-ops on an empty
// library and the player loses at the next SBA per CR 704.5b —
// exactly what "drawing from an empty library" does in paper.
func init() {
	Register(Spec{
		OracleID:     "53236dd7-845a-444c-96d5-f41ed7325d8f",
		Name:         "Rhystic Study",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The draw is mandatory when the opponent declines to pay — you can't choose to skip it, which matters on an empty library."},
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
								return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
							},
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
