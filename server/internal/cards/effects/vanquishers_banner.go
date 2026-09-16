package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vanquisher's Banner — Artifact for {5}:
//
//	"As this artifact enters, choose a creature type.
//	 Creatures you control of the chosen type get +1/+1.
//	 Whenever you cast a creature spell of the chosen type, draw a
//	 card."
//
// Door of Destinies' sibling with a flat anthem and a draw instead of
// a counter. The one difference in the trigger is load-bearing: this
// one says "a CREATURE spell of the chosen type", so a Kindred
// Instant naming the tribe grows a Door and draws nothing here.
func init() {
	Register(Spec{
		OracleID: "8cf38025-5821-45a0-9483-266353b7e82d",
		Name:     "Vanquisher's Banner",
		AsEnters: ChooseCreatureTypeAsEnters("Vanquisher's Banner"),
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Chosen: true, YoursOnly: true}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller || source.NamedTribe == "" {
					return false
				}
				spell, ok := g.LookupCardForEffect(ev.CardID)
				return ok && spell.IsCreature() && spell.HasSubtype(source.NamedTribe)
			}, "Vanquisher's Banner — draw a card", func(g *game.Game, item *game.StackItem) error {
				return g.DrawNForEffect(item.Controller, 1)
			}),
		},
	})
}
