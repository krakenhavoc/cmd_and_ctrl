package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sylvan Echoes — Enchantment {G} (Lorwyn, S82 OLD b546, #5593):
//
//	"Whenever you clash and win, you may draw a card. (This ability
//	 triggers after the clash ends.)"
//
// Marvo's clash watcher (game.EventClashWon, the winner as Actor) with an
// optional draw (CR 603.5): the question is asked as the trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6d19e2bb-2e5f-4e55-bb86-de01ea35263d",
		Name:         "Sylvan Echoes",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Optional(
				effects.On(game.EventClashWon, effects.YouWonAClash, "Sylvan Echoes — you may draw a card", effects.Do(effects.DrawCards{N: 1})),
				"Sylvan Echoes — draw a card?"),
				game.Purpose{Draws: 1}),
		},
	})
}
