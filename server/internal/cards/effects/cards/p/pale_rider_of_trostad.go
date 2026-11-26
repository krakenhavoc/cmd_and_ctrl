package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pale Rider of Trostad — Creature — Spirit {1}{B}, 3/3 (Shadows over
// Innistrad, #5543):
//
//	"Skulk (This creature can't be blocked by creatures with greater
//	 power.)
//	 When this creature enters, discard a card."
//
// Skulk is the engine keyword (CR 702.118). The discard is its controller's
// own choice, Bloodrage Brawler's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6cc47c8e-9154-4517-940d-d1a7f35e5677",
		Name:            "Pale Rider of Trostad",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"skulk"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Pale Rider of Trostad — discard a card", func(g *game.Game, item *game.StackItem) error {
				return effects.DiscardOne(g, item, item.Controller)
			}),
		},
	})
}
