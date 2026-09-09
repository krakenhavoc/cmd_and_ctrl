package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Magmakin Artillerist — 1/4 Creature — Elemental Pirate for {2}{R}:
//
//	"Whenever you discard one or more cards, this creature deals that
//	 much damage to each opponent.
//	 Cycling {1}{R}. When you cycle this card, it deals 1 damage to
//	 each opponent."
//
// Glint-Horn Buccaneer's effect on a body that doesn't need to
// attack. With the commander looting every turn this is a slow but
// inevitable clock on the whole table.
//
// Batching (CR 603.1): the card reads "one or more cards… that
// much damage", one trigger per batch. The engine emits one
// EventDiscardCard per card, so a two-card discard deals 1 twice
// rather than 2 once. Same total, two log lines.
//
// Cycling is an alternative cast path (S29) and isn't modelled, so
// the cycling trigger can't fire either.
func init() {
	Register(Spec{
		OracleID: "900b9409-9c16-414d-8674-2ea42c2415a1",
		Name:     "Magmakin Artillerist",
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDiscardCard},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return discardedByYou(ev, source)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Magmakin Artillerist — 1 damage to each opponent",
					func(g *game.Game, item *game.StackItem) error {
						return damageToEachOpponent(g, item, 1)
					})
			},
		}},
	})
}
