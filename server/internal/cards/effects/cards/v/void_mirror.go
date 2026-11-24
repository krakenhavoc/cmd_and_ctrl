package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Void Mirror — Artifact {2} (Modern Horizons 2, #5501):
//
//	"Whenever a player casts a spell, if no colored mana was spent to cast
//	 it, counter that spell."
//
// Boromir's read of the cast's payment, for every player's casts: the
// spell's recorded payment shows no colored mana (colorless mana, or none
// at all, as a free cast). A payment the engine did not record is not
// treated as colorless. The "if" is read again on resolution (CR 603.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "08aaafc7-5fc0-4c28-a361-2a22536ec080",
		Name:           "Void Mirror",
		Completeness:   effects.CompletenessFull,
		ReadsManaSpent: game.ReadsEveryCast,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Kind == game.EventCast && effects.S82OldB39NoColoredManaWasSpent(g, ev.CardID)
			}, "Void Mirror — counter that spell (no colored mana was spent)", effects.S82OldB39CounterTheSpellIfNoColoredManaWasSpent),
		},
	})
}
