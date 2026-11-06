package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "41c15160-ee8c-42e3-bc6c-593b8c8ae335",
		Name:         "Crash Through",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		OnResolve: ResolveSteps(
			GrantKeywordUntilEOT{
				Match:    And(Creature(), YouControl()),
				Keywords: []string{"trample"},
				Label:    "Crash Through — creatures you control gain trample",
			},
			DrawCards{N: 1},
		),
	})
}
