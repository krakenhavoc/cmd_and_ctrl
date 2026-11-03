package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Archon of Redemption — Creature — Archon {3}{W}{W}, 3/4 (Jumpstart):
//
//	"Flying
//	 Whenever this creature or another creature you control with flying
//	 enters, you may gain life equal to that creature's power."
//
// One trigger with two ways in: the Archon's own entry, and the entry of
// any other creature with flying under its controller's control, judged as
// it entered. The power is the entered creature's as the trigger resolves
// (CR 608.2h): its current power while it is still the same object on the
// battlefield, so a pump in response counts, and the power it last had
// there once it has left. A negative power gains nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "434e58f4-4237-454f-a2d0-61885c9aafaa",
		Name:            "Archon of Redemption",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.CardID == source.InstanceID {
					return true
				}
				c, ok := enteredUnderYourControl(ev, source, g, true)
				return ok && c.IsCreature() && game.HasKeyword(&c, "flying")
			}, "Archon of Redemption — you may gain life equal to that creature's power", func(g *game.Game, item *game.StackItem) error {
				o := item.Trigger.Object
				if o == nil {
					return nil
				}
				info, ok := g.PermanentForEffect(game.ObjectRef{ID: o.ID, Epoch: o.Epoch})
				if !ok || info.Power <= 0 {
					return nil
				}
				return GainLife{Amount: info.Power}.Apply(NewContext(g, item))
			}), "Gain life equal to that creature's power?"),
		},
	})
}
