package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// burningTreeVandalLabel is the attack trigger's stack label and its
// discard prompt.
const burningTreeVandalLabel = "Burning-Tree Vandal — you may discard a card; if you do, draw a card"

// Burning-Tree Vandal — Creature — Human Rogue {2}{R}, 2/1:
//
//	"Riot (This creature enters with your choice of a +1/+1 counter or
//	 haste.)
//	 Whenever this creature attacks, you may discard a card. If you do,
//	 draw a card."
//
// Riot is the engine's keyword (ADR 0109 §10). The rummage is Cool but
// Rude's: the discard comes first, from the hand as it is, and the draw
// is the number actually discarded, so declining draws nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5bed3702-984d-436f-bb71-e16eaeffe902",
		Name:            "Burning-Tree Vandal",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordRiot},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks(burningTreeVandalLabel, func(g *game.Game, item *game.StackItem) error {
				return b39MayDiscardThenDraw(1, false, burningTreeVandalLabel,
					func(discarded int) int { return discarded })(NewContext(g, item))
			}),
		},
	})
}
