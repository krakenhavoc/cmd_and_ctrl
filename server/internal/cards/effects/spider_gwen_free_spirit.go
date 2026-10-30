package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spider-Gwen, Free Spirit — Legendary Creature — Spider Human Hero
// {2}{R}, 2/3 (Marvel Super Heroes Commander, #4455):
//
//	"Reach
//	 Whenever Spider-Gwen becomes tapped, you may discard a card. If you
//	 do, draw a card."
//
// Tapping for an attack and tapping for a cost both count
// (ThisBecameTapped, Wakandan Tusker's reader); the discard is optional
// and the draw happens only if a card was discarded (the shared
// may-discard-then-draw prompt of Battlefield Scavenger).
//
// No simplification.
func init() {
	const label = "Spider-Gwen, Free Spirit — you may discard a card; if you do, draw a card"
	Register(Spec{
		OracleID:        "2648559d-1d96-47d2-a769-cfa2cbca2be5",
		Name:            "Spider-Gwen, Free Spirit",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, ThisBecameTapped, label,
				func(g *game.Game, item *game.StackItem) error {
					return b39MayDiscardThenDraw(1, false, label, drawAsManyAsDiscarded)(NewContext(g, item))
				}),
		},
	})
}
