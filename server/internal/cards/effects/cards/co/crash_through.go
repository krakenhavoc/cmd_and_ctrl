package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crash Through — Sorcery {R} (Foundations, #4465):
//
//	"Creatures you control gain trample until end of turn.
//	 Draw a card."
//
// Blinding Fog's grant: the creatures you control as the spell resolves
// gain trample, and one that enters later does not (CR 611.2c). The
// draw follows.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "41c15160-ee8c-42e3-bc6c-593b8c8ae335",
		Name:         "Crash Through",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		OnResolve: effects.ResolveSteps(
			effects.GrantKeywordUntilEOT{
				Match:    effects.And(effects.Creature(), effects.YouControl()),
				Keywords: []string{"trample"},
				Label:    "Crash Through — creatures you control gain trample",
			},
			effects.DrawCards{N: 1},
		),
	})
}
