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
// Sandbox simplification, declared, the Wrathful Red Dragon posture:
// the engine emits one damage event per SOURCE, so a creature blocked
// by two attackers fires twice — once per blocker's damage — where
// the printed card fires once for the total. The same damage lands
// either way; only the split differs.
func init() {
	Register(Spec{
		OracleID:     "f6069a4f-744e-43e8-9f8b-7c13b8b0187f",
		Name:         "Repercussion",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"If two or more sources damage a creature at the same time, its controller takes each source's damage separately instead of the total in one go."},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := b28CreatureWasDealtDamage(ev, g)
				return ok
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				c, ok := b28CreatureWasDealtDamage(ev, g)
				if !ok {
					return nil
				}
				victim, amount := c.Controller, ev.Amount
				return game.NewTriggeredItem(source, "Repercussion — that much damage to the creature's controller",
					func(g *game.Game, item *game.StackItem) error {
						return DealDamage{Source: item.SourceCardID, Target: victim, Amount: amount}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
