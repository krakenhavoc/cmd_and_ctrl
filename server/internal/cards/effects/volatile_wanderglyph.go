package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Volatile Wanderglyph — Artifact Creature — Golem {1}{R}, 2/2 (The Lost
// Caverns of Ixalan, #4470):
//
//	"Whenever this creature becomes tapped, you may discard a card. If you
//	 do, draw a card."
//
// Attentive Sunscribe's trigger: "becomes tapped" is the tap event and
// the attack that taps it (ThisBecameTapped, #4534), so a vigilance
// attack is not one, and any other tap counts (CR 701.26a). The rummage
// is Rubble Rouser's: the discard first, and a draw only for a card
// actually discarded; declining draws nothing.
//
// No simplification.
func init() {
	const label = "Volatile Wanderglyph — you may discard a card; if you do, draw a card"
	Register(Spec{
		OracleID:     "ebc3ca0f-35eb-4606-90ab-eb64556f81fe",
		Name:         "Volatile Wanderglyph",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, ThisBecameTapped, label,
				func(g *game.Game, item *game.StackItem) error {
					return b39MayDiscardThenDraw(1, false, label, drawOneIfDiscarded)(NewContext(g, item))
				}),
		},
	})
}
