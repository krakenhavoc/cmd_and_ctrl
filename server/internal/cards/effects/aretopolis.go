package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aretopolis — Plane — Kephalai (#4870):
//
//	"When you planeswalk to Aretopolis and at the beginning of your upkeep,
//	 put a scroll counter on Aretopolis, then you gain life equal to the
//	 number of scroll counters on it.
//	 When Aretopolis has ten or more scroll counters on it, planeswalk.
//	 Whenever chaos ensues, put a scroll counter on Aretopolis, then draw cards
//	 equal to the number of scroll counters on it."
//
// The counters sit on the face-up planar card
// (game.PutCountersOnFaceUpPlanarCardForEffect) and every scroll counter
// counts, whoever's ability put it there (the ruling). Putting counters on it
// announces the new total (EventCounterPlaced), which is what the third
// ability watches: the counters arrive one at a time, so the trigger fires
// as the total reaches ten (CR 603.8's state trigger, once per crossing). A
// planeswalk away makes the plane a new object with no counters (CR 901.7a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d6ab4159-e04e-4991-8a9b-9da302c98e9d",
		Name:         "Aretopolis",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenYouPlaneswalkToThisAndAtYourUpkeep("Aretopolis — put a scroll counter on it, then you gain life equal to the number of scroll counters on it",
				func(g *game.Game, item *game.StackItem) error {
					n, ok := g.PutCountersOnFaceUpPlanarCardForEffect(item.SourceCardID, aretopolisScroll, 1)
					if !ok {
						return nil
					}
					return GainLife{Player: item.Controller, Amount: n}.Apply(NewContext(g, item))
				}),
			On(game.EventCounterPlaced, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Target == source.InstanceID && ev.Label == aretopolisScroll && ev.Amount == 10
			}, "Aretopolis — planeswalk", func(g *game.Game, item *game.StackItem) error {
				if g.Planechase == nil {
					return nil
				}
				return g.PlaneswalkForEffect(item.Controller)
			}),
			WheneverChaosEnsues("Aretopolis — put a scroll counter on it, then draw cards equal to the number of scroll counters on it",
				func(g *game.Game, item *game.StackItem) error {
					n, ok := g.PutCountersOnFaceUpPlanarCardForEffect(item.SourceCardID, aretopolisScroll, 1)
					if !ok {
						return nil
					}
					return DrawCards{Player: item.Controller, N: n}.Apply(NewContext(g, item))
				}),
		},
	})
}

// aretopolisScroll is the counter's name.
const aretopolisScroll = "scroll"
