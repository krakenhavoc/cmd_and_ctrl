package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Drelnoch — Creature — Yeti Mutant {4}{U}, 3/3 (Coldsnap):
//
//	"Whenever this creature becomes blocked, you may draw two cards."
//
// It triggers once each time the Drelnoch becomes blocked, however many
// creatures block it (CR 509.3c); "you may" is asked as the trigger resolves
// (CR 603.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c25465a9-a0f3-48be-8bdf-4aa861da7d3e",
		Name:         "Drelnoch",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventBecomesBlocked, effects.Self, "Drelnoch — you may draw two cards",
				effects.Do(effects.DrawCards{N: 2})),
				"Drelnoch — draw two cards?"),
		},
	})
}
