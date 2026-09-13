package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Prized Statue — Artifact {2} (EDHREC rank 2313):
//
//	"When this artifact enters or is put into a graveyard from the
//	 battlefield, create a Treasure token. (It's an artifact with
//	 "{T}, Sacrifice this token: Add one mana of any color.")"
//
// A two-mana rock that refunds itself twice. One printed ability
// with two trigger conditions (CR 603.2c), so one TriggeredAbility
// watching two kinds: its own entry (EventETB) and its own death
// (EventLTB into a graveyard — cardDied, so a bounce or an exile
// makes nothing). The dies half is harvested off the card in the
// graveyard with its controller intact, so the Treasure goes to
// whoever controlled the Statue as it died.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "681fc668-cb26-4ba4-a915-48ddfa2b9520",
		Name:         "Prized Statue",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB, game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				if ev.Kind == game.EventETB {
					return ev.CardID == source.InstanceID
				}
				return cardDied(ev, source)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Prized Statue — create a Treasure",
					func(g *game.Game, item *game.StackItem) error {
						return CreateToken{Controller: item.Controller, Template: TreasureToken(), N: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
