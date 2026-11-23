package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "2648559d-1d96-47d2-a769-cfa2cbca2be5",
		Name:            "Spider-Gwen, Free Spirit",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, effects.ThisBecameTapped, label,
				func(g *game.Game, item *game.StackItem) error {
					return effects.B39MayDiscardThenDraw(1, false, label, effects.DrawAsManyAsDiscarded)(effects.NewContext(g, item))
				}),
		},
	})
}
