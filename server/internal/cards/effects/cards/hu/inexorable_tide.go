package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inexorable Tide — Enchantment {3}{U}{U}:
//
//	"Whenever you cast a spell, proliferate."
//
// Flux Channeler with no creature clause and no body: EVERY spell,
// creatures included, and nothing can attack it down. Five mana that
// does nothing the turn it lands and wins the game two turns later
// in a deck built around it.
//
// It triggers on its own cast? No — the Tide is not on the
// battlefield while it is being cast, so there is nothing to
// trigger. The first spell it sees is the next one.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "40ec0a47-badf-4074-b0a8-749bb7c17b95",
		Name:         "Inexorable Tide",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.ByYou, "Inexorable Tide — proliferate", effects.Do(effects.Proliferate{})),
		},
	})
}
