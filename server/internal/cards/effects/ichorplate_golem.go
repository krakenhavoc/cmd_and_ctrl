package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ichorplate Golem — Artifact Creature — Phyrexian Golem {3}, 2/3
// (Phyrexia: All Will Be One):
//
//	"Whenever a creature you control enters, if it has one or more oil
//	 counters on it, put an oil counter on it.
//	 Creatures you control with oil counters on them get +1/+1."
//
// The trigger is checked as the creature enters and again as it resolves
// (CR 603.4); a creature that has left is not counted. The anthem is a
// layer 7c modifier read live, so a creature gaining or losing its oil
// counters changes at once. The Golem counts itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2feca0cd-6bd3-4e19-8cfd-81d7bd1118ee",
		Name:         "Ichorplate Golem",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "Creatures you control with oil counters on them get +1/+1.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller && target.Counters[s82OneOil] > 0
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power++
				c.Toughness++
			},
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				entered, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && entered.IsCreature() && entered.Counters[s82OneOil] > 0
			}, "Ichorplate Golem — put an oil counter on that creature", func(g *game.Game, item *game.StackItem) error {
				if item.Trigger == nil {
					return nil
				}
				id := item.Trigger.Event.CardID
				c, ok := g.LookupCardForEffect(id)
				if !ok || !onBattlefield(g, id) || c.Counters[s82OneOil] == 0 {
					return nil
				}
				return AddCounter{Target: id, Kind: s82OneOil, N: 1}.Apply(NewContext(g, item))
			}),
		},
	})
}
