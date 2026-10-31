package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Krang, Master Mind — {6}{U}{U} Legendary Artifact Creature — Utrom Warrior
// 1/4 (Teenage Mutant Ninja Turtles, #4457):
//
//	"Affinity for artifacts (This spell costs {1} less to cast for each
//	 artifact you control.)
//	 When Krang enters, if you have fewer than four cards in hand, draw
//	 cards equal to the difference.
//	 Krang gets +1/+0 for each other artifact you control."
//
// Affinity is Myr Enforcer's self cost modifier (CR 702.41a). The enters
// trigger is an intervening "if" (CR 603.4): it triggers only with fewer than
// four cards in hand, is checked again as it resolves, and draws the
// difference counted then (The Ten Rings' AtYourEndStepDrawUpTo, on an
// enters trigger). The pump is Iron Man, Master of Machines' layer 7c count
// of the other artifacts its controller controls, Krang himself excluded.
//
// No simplification.
func init() {
	const label = "Krang, Master Mind — draw cards equal to the difference between four and your hand size"
	Register(Spec{
		OracleID:     "52562194-4a97-40d2-aa73-7827f7834f2f",
		Name:         "Krang, Master Mind",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			AffinityFor("Affinity for artifacts", Artifact()),
		},
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			Label:    "Krang gets +1/+0 for each other artifact you control.",
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power += permanentsYouControl(g, source, func(other *game.Card) bool {
					return other.InstanceID != source.InstanceID && other.IsArtifact()
				})
			},
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.CardID == source.InstanceID && b14HandSize(g, source.Controller) < 4
			}, label, func(g *game.Game, item *game.StackItem) error {
				short := 4 - b14HandSize(g, item.Controller)
				if short <= 0 {
					return nil
				}
				return DrawCards{Player: item.Controller, N: short}.Apply(NewContext(g, item))
			}),
		},
	})
}
