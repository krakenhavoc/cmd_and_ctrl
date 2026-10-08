package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kolodin, Triumph Caster — Legendary Creature — Human Pilot {R}{W}:
//
//	"Mounts and Vehicles you control have haste.
//	 Whenever a Mount you control enters, it becomes saddled until end
//	 of turn.
//	 Whenever a Vehicle you control enters, it becomes an artifact
//	 creature until end of turn."
//
// The entering permanent is recorded when the trigger is built, with its
// object epoch, so a Mount that has left and come back is a new object
// the old trigger does not saddle (#1432). Kolodin is not a Mount or a
// Vehicle, so he triggers for neither of his own entries.
//
// No simplification.
func init() {
	mountsAndVehicles := func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.Controller == source.Controller && (target.HasSubtype("Mount") || target.HasSubtype("Vehicle"))
	}
	enters := func(subtype, label string, effect Effect) game.TriggeredAbility {
		return game.TriggeredAbility{
			Watches: []game.EventKind{game.EventETB},
			Key:     label,
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, true)
				return ok && c.HasSubtype(subtype)
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, label)
				epoch := -1
				if c, ok := g.LookupCardForEffect(ev.CardID); ok {
					epoch = c.ObjectEpoch
				}
				item.Params.Object = game.ObjectRef{ID: ev.CardID, Epoch: epoch}
				return item
			},
			Effect: effect,
		}
	}
	// stillThatObject resolves the recorded permanent, if it is the same
	// object and still on the battlefield.
	stillThatObject := func(g *game.Game, item *game.StackItem) (game.ObjectRef, bool) {
		o := item.Params.Object
		c, ok := g.LookupCardForEffect(o.ID)
		return o, ok && c.ObjectEpoch == o.Epoch && onBattlefield(g, o.ID)
	}
	Register(Spec{
		OracleID:     "45682055-d02a-450c-a41d-aa67162ff015",
		Name:         "Kolodin, Triumph Caster",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{KeywordGrant(mountsAndVehicles, "haste")},
		Triggered: []game.TriggeredAbility{
			enters("Mount", "Kolodin, Triumph Caster — that Mount becomes saddled until end of turn", func(g *game.Game, item *game.StackItem) error {
				o, ok := stillThatObject(g, item)
				if !ok {
					return nil
				}
				return BecomeSaddled{Target: o.ID}.Apply(NewContext(g, item))
			}),
			enters("Vehicle", "Kolodin, Triumph Caster — that Vehicle becomes an artifact creature until end of turn", func(g *game.Game, item *game.StackItem) error {
				o, ok := stillThatObject(g, item)
				if !ok {
					return nil
				}
				return BecomeCreatureUntilEOT{Target: o.ID, Label: "Kolodin, Triumph Caster — becomes an artifact creature"}.Apply(NewContext(g, item))
			}),
		},
	})
}
