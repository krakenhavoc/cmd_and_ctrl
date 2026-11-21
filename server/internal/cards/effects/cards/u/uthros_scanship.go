package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Uthros Scanship — Artifact — Spacecraft {3}{U}, 4/4 (Edge of
// Eternities, #4461):
//
//	"When this Spacecraft enters, draw two cards, then discard a card.
//	 Station
//	 8+ | Flying"
//
// The enters trigger is the shared loot body (b16DrawThenDiscard): the
// discard is chosen after both draws. Station and the 8+ line are
// Susurian Dirgecraft's helpers: at eight or more charge counters it is a
// 4/4 artifact creature with flying (CR 702.184).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7b4c37dc-8cb0-4870-929e-11c2a45952a2",
		Name:         "Uthros Scanship",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Uthros Scanship — draw two cards, then discard a card",
				func(g *game.Game, item *game.StackItem) error { return effects.B16DrawThenDiscard(g, item, 2) }),
		},
		Activated: []effects.ActivatedAbility{effects.Station()},
		Static: append(
			effects.SpacecraftAt(8, 4, 4),
			effects.ThresholdKeywords(8, "flying"),
		),
	})
}
