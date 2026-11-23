package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ravenous Necrotitan — Creature — Phyrexian Horror {2}{B}{B}, 6/6:
//
//	"Corrupted — When this creature enters, sacrifice a creature unless an
//	 opponent has three or more poison counters."
//
// Not an intervening "if": the trigger always goes on the stack and the
// poison is read as it resolves (CR 608.2c), so an opponent reaching three
// counters in response saves the sacrifice. The Necrotitan is itself a
// legal choice (Accursed Centaur).
//
// No simplification.
func init() {
	const label = "Corrupted — Ravenous Necrotitan — sacrifice a creature unless an opponent has three or more poison counters"
	effects.Register(effects.Spec{
		OracleID:     "90109cb4-1e16-4d54-9332-b5fb1e71042c",
		Name:         "Ravenous Necrotitan",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
				if effects.SeedcoreCorrupted(g, item.Controller, item.SourceCardID) {
					return nil
				}
				g.PlayerSacrificesForEffect(item.SourceCardID, item.Controller, effects.SacrificeSpec("a creature", effects.Creature()), label)
				return nil
			}),
		},
	})
}
