package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Exuberant Fuseling — Creature — Phyrexian Goblin Warrior {R}, 0/1
// (Phyrexia: All Will Be One):
//
//	"Trample
//	 This creature gets +1/+0 for each oil counter on it.
//	 When this creature enters and whenever another creature or artifact
//	 you control is put into a graveyard from the battlefield, put an oil
//	 counter on this creature."
//
// One printed ability with two conditions: it watches both the entry and
// the departures. The departure half is Vat of Rebirth's reading.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "754e5238-00a2-45fa-9b20-40980b73f1cc",
		Name:            "Exuberant Fuseling",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{
			s82OilPump(1, 0, "This creature gets +1/+0 for each oil counter on it."),
		},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventETB, game.EventLTB},
				func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
					if ev.Kind == game.EventETB {
						return Self(ev, source, lki, g)
					}
					return s82OilOnDeparture(ev, source, lki, g)
				},
				"Exuberant Fuseling — put an oil counter on it",
				Do(CounterOnThis{Kind: s82OneOil, N: 1})),
		},
	})
}
