package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Door of Destinies — Artifact for {4}:
//
//	"As this artifact enters, choose a creature type.
//	 Whenever you cast a spell of the chosen type, put a charge
//	 counter on this artifact.
//	 Creatures you control of the chosen type get +1/+1 for each
//	 charge counter on this artifact."
//
// Three clauses, three different mechanisms, which is why it is the
// second card in this sprint rather than the tenth: the CR 614.12
// choice, a cast trigger keyed to that choice, and a scaling anthem
// that reads counters off its own source.
//
// The trigger says "a spell of the chosen type", not "a CREATURE
// spell" — Kindred (Tribal) cards count, and a Kindred Instant naming
// the chosen type really does grow the Door. HasSubtype on the stack
// object is what gets that right, because a spell on the stack has no
// effective characteristic and falls through to its printed type
// line, where the Kindred subtype lives.
func init() {
	Register(Spec{
		OracleID: "9b6d3dcf-aa5a-4516-bd48-a0723e86bfd1",
		Name:     "Door of Destinies",
		AsEnters: ChooseCreatureTypeAsEnters("Door of Destinies"),
		Static: []game.StaticAbility{
			TribalScalingAnthem(
				TribeFilter{Chosen: true, YoursOnly: true},
				func(source *game.Card, _ *game.Game) int {
					return source.Counters["charge"]
				},
			),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller || source.NamedTribe == "" {
					return false
				}
				spell, ok := g.LookupCardForEffect(ev.CardID)
				return ok && spell.HasSubtype(source.NamedTribe)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				sourceID := source.InstanceID
				return game.NewTriggeredItem(source, "Door of Destinies — put a charge counter",
					func(g *game.Game, _ *game.StackItem) error {
						return g.AddCounterForEffect(sourceID, "charge", 1)
					})
			},
		}},
	})
}
