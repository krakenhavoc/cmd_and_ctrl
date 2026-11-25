package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nath of the Gilt-Leaf — Legendary Creature — Elf Warrior {3}{B}{G}, 4/4
// (Commander 2016, S82 older sets, #5537):
//
//	"At the beginning of your upkeep, you may have target opponent discard a
//	 card at random.
//	 Whenever an opponent discards a card, you may create a 1/1 green Elf
//	 Warrior creature token."
//
// The opponent is the upkeep trigger's target and the "may" is answered as it
// resolves; each card an opponent discards, by Nath's own ability or any
// other, is one separate trigger with its own "may" (CR 603.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "87a44a87-a2a5-4193-a163-e4846e5b8387",
		Name:         "Nath of the Gilt-Leaf",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.AtYourUpkeep("Nath of the Gilt-Leaf — you may have target opponent discard a card at random",
				effects.Do(effects.TargetPlayerDiscardsAtRandom{N: 1})),
				effects.TargetPlayer("target opponent", effects.Opponent())),
				"Nath of the Gilt-Leaf — have target opponent discard a card at random?"),
			effects.Optional(effects.On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.B18OpponentDiscarded(ev, source)
			},
				"Nath of the Gilt-Leaf — you may create a 1/1 green Elf Warrior creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green Elf Warrior"), N: 1})),
				"Nath of the Gilt-Leaf — create a 1/1 green Elf Warrior creature token?"),
		},
	})
}
