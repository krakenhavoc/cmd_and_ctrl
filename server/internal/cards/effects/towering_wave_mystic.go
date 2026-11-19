package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Towering-Wave Mystic — Creature — Merfolk Wizard {1}{U}, 2/1 (Foundations Jumpstart):
//
//	"Whenever this creature deals damage, target player mills that many
//	 cards."
//
// Horned Cheetah's trigger: any damage this creature deals, combat or not,
// read off the event's Amount (the damage actually dealt, after
// prevention). The target player is chosen as the trigger goes on the stack
// and may be any player, the controller included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "121a1efa-03f2-4b5f-8e72-539de028457e",
		Name:         "Towering-Wave Mystic",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(game.TriggeredAbility{
				Watches: []game.EventKind{game.EventDealDamage},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.Amount > 0 && ev.Source == source.InstanceID
				},
				Key: "Towering-Wave Mystic — target player mills that many cards",
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetPlayer {
							return MillCards{Player: t.ID, N: item.Trigger.Event.Amount}.Apply(ctx)
						}
					}
					return nil
				},
			}, TargetPlayer("target player")),
		},
	})
}
