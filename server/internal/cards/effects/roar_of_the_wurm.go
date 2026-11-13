package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Roar of the Wurm — Sorcery {6}{G} (Odyssey, #5613, #5789):
//
//	"Create a 6/6 green Wurm creature token.
//	 Flashback {3}{G} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// Call of the Herd's shape with the table's plain 6/6 green Wurm row.
// Flashback is the shared constructor (graveyard cast, price, and CR
// 702.34a's exile as the spell leaves the stack).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "b8b9e6fd-b3ed-4fbc-8753-74018fa33caa",
		Name:             "Roar of the Wurm",
		Completeness:     CompletenessFull,
		Purpose:          game.Purpose{Tokens: 1, CreatureTokens: 1},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{3}{G}")},
		OnResolve:        ResolveSteps(CreateToken{Template: TokenCard("6/6 green Wurm"), N: 1}),
	})
}
