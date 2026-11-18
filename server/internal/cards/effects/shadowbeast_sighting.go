package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shadowbeast Sighting — Sorcery {3}{G} (Midnight Hunt, #5497):
//
//	"Create a 4/4 green Beast creature token.
//	 Flashback {6}{G} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// The table's plain 4/4 green Beast row, and the shared Flashback
// constructor (graveyard cast, price, CR 702.34a's exile as the spell
// leaves the stack).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "f05d5632-0a82-45bf-b66e-22cf4080ed5e",
		Name:             "Shadowbeast Sighting",
		Completeness:     CompletenessFull,
		Purpose:          game.Purpose{Tokens: 1, CreatureTokens: 1},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{6}{G}")},
		OnResolve:        ResolveSteps(CreateToken{Template: TokenCard("4/4 green Beast"), N: 1}),
	})
}
