package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chalk Outline — Enchantment {3}{G} (Murders at Karlov Manor, #4469):
//
//	"Whenever one or more creature cards leave your graveyard, create a
//	 2/2 white and blue Detective creature token, then investigate.
//	 (Create a Clue token. It's an artifact with '{2}, Sacrifice this
//	 token: Draw a card.')"
//
// Desecrated Tomb's batch trigger: a mass return makes one Detective
// and one Clue, not one of each per card, and the creature cards may
// leave in any direction. The Detective is made first and the Clue
// after it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "af335eb3-f60a-44b3-a1ed-1e165f3e3605",
		Name:         "Chalk Outline",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.OnAny(effects.LeftGraveyardEvents,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.CreatureCardLeftYourGraveyard(ev, source, g)
				},
				"Chalk Outline — create a 2/2 Detective, then investigate",
				effects.Do(
					effects.CreateToken{Template: effects.TokenCard("2/2 white and blue Detective"), N: 1},
					effects.Investigate{N: 1},
				))),
		},
	})
}
