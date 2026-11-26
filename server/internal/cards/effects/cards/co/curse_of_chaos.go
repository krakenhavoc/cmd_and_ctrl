package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Curse of Chaos — Enchantment — Aura Curse {2}{R} (Commander 2013, S82 older
// sets):
//
//	"Enchant player
//	 Whenever a player attacks enchanted player with one or more creatures,
//	 that attacking player may discard a card. If the player does, they draw
//	 a card."
//
// One trigger per declaration per attacking player, controlled by that
// player. The discard is chosen as it resolves and the draw happens only if
// a card was discarded (Bitter Reunion's may-discard-then-draw).
//
// No simplification.
func init() {
	const label = "Curse of Chaos — the attacking player may discard a card; if they do, they draw a card"
	effects.Register(effects.Spec{
		OracleID:     "a242af4b-5de6-4961-b684-f4bd809977c6",
		Name:         "Curse of Chaos",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPlayer(),
		Triggered: []game.TriggeredAbility{
			effects.OldB383WhenAPlayerAttacksEnchantedPlayer(label, func(g *game.Game, item *game.StackItem) error {
				return effects.B39MayDiscardThenDraw(1, false, "Curse of Chaos — discard a card? If you do, draw a card.", effects.DrawOneIfDiscarded)(effects.NewContext(g, item))
			}),
		},
	})
}
