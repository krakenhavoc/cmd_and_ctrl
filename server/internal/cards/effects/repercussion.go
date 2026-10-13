package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Repercussion — Enchantment {1}{R}{R} (EDHREC rank 2983):
//
//	"Whenever a creature is dealt damage, this enchantment deals that
//	 much damage to that creature's controller."
//
// The Blasphemous Act combo piece. The trigger watches
// EventDealDamage for ANY creature in the Target slot — anyone's,
// the controller's own included, combat damage and burn alike
// (b28CreatureWasDealtDamage) — reads the creature's controller live
// (the event fires before the state-based sweep, so a creature that
// took lethal is still there to be read), and Repercussion deals the
// event's amount to that player on resolution: red noncombat damage
// from an enchantment, as printed.
//
// The engine emits one damage event per SOURCE, so the trigger waits for
// the damage to finish (AtBatchEnd, ThatMuchDamage) and deals the creature's controller the
// TOTAL the creature was dealt at the same time, as printed (CR 120.3, CR 510.2).
func init() {
	Register(Spec{
		OracleID:     "f6069a4f-744e-43e8-9f8b-7c13b8b0187f",
		Name:         "Repercussion",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := b28CreatureWasDealtDamage(ev, g)
				return ok
			},
			Key:        "Repercussion — that much damage to the creature's controller",
			AtBatchEnd: ThatMuchDamage(),
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				c, ok := b28CreatureWasDealtDamage(ev, g)
				if !ok {
					return nil
				}
				item := game.NewTriggeredItem(source, "Repercussion — that much damage to the creature's controller")
				item.Params.Player = c.Controller
				item.Params.Amount = ev.Amount
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DealDamage{Source: item.SourceCardID, Target: item.Params.Player, Amount: item.Params.Amount}.Apply(NewContext(g, item))
			},
		}},
	})
}
