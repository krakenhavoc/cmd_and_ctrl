package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Evolutionary Leap — Enchantment {1}{G}:
//
//	"{G}, Sacrifice a creature: Reveal cards from the top of your library
//	 until you reveal a creature card. Put that card into your hand and
//	 the rest on the bottom of your library in a random order."
//
// The sacrifice is a cost. The run is revealed and the creature card taken
// (Abundant Harvest's socB30RevealUntilToHandRestBottom); with no hit the
// whole library goes under in a random order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "44cf5ef4-2b0f-4d9d-bf87-8835424721d0",
		Name:         "Evolutionary Leap",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}, Sacrifice a creature: Reveal cards from the top of your library until you reveal a creature card. Put that card into your hand and the rest on the bottom of your library in a random order.",
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{G}"), effects.SacrificeACreature()),
			Effect: effects.SocB30RevealUntilToHandRestBottom("Evolutionary Leap — creature card",
				func(c game.Card) bool { return c.IsCreature() }),
		}},
	})
}
