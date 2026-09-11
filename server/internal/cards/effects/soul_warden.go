package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Soul Warden — Creature — Human Cleric {W}, 1/1 (EDHREC rank 488):
//
//	"Whenever another creature enters, you gain 1 life."
//
// ANY creature, under anyone's control — the Warden gains off an
// opponent's Krenko activation as happily as off your own. "Another"
// excludes its own entry only. One trigger per creature, so a
// five-token batch is five triggers, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f3fad295-1af2-4ecc-8546-b121ad6be27b",
		Name:         "Soul Warden",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.CardID == source.InstanceID {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature()
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Soul Warden — you gain 1 life",
					func(g *game.Game, item *game.StackItem) error {
						return GainLife{Player: item.Controller, Amount: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
