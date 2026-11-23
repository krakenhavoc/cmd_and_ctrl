package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "754e5238-00a2-45fa-9b20-40980b73f1cc",
		Name:            "Exuberant Fuseling",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{
			effects.S82OilPump(1, 0, "This creature gets +1/+0 for each oil counter on it."),
		},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventETB, game.EventLTB},
				func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
					if ev.Kind == game.EventETB {
						return effects.Self(ev, source, lki, g)
					}
					return effects.S82OilOnDeparture(ev, source, lki, g)
				},
				"Exuberant Fuseling — put an oil counter on it",
				effects.Do(effects.CounterOnThis{Kind: effects.S82OneOil, N: 1})),
		},
	})
}
