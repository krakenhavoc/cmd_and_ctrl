package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ichor Wellspring — Artifact for {2}:
//
//	"When Ichor Wellspring enters the battlefield or is put into a
//	graveyard from the battlefield, draw a card."
//
// One card, two triggers on the same ability — the engine models
// that as two entries watching different events, which is how the
// oracle text reads anyway ("enters OR is put into a graveyard").
//
// The dies half fires on ANY route to the graveyard, sacrifice
// included, which is the point: paired with a sacrifice outlet the
// Wellspring is a two-card draw engine. It's also the first
// non-creature dies-trigger in the catalog, so it exercises the
// LTB path for a permanent that was never a creature.
func init() {
	drawOne := func(label string) game.TriggeredAbility {
		return game.TriggeredAbility{
			Key: label,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			},
		}
	}
	etb := drawOne("Ichor Wellspring — draw a card (entered)")
	etb.Watches = []game.EventKind{game.EventETB}
	etb.AppliesTo = func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return ev.CardID == source.InstanceID
	}
	dies := drawOne("Ichor Wellspring — draw a card (died)")
	dies.Watches = []game.EventKind{game.EventLTB}
	dies.AppliesTo = func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return cardDied(ev, source)
	}
	Register(Spec{
		OracleID:     "5b5ef43b-13fd-4461-8d2d-18be65e9a790",
		Name:         "Ichor Wellspring",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{etb, dies},
	})
}
