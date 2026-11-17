package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Call of the Herd — Sorcery {2}{G} (Odyssey, #5613):
//
//	"Create a 3/3 green Elephant creature token.
//	 Flashback {3}{G} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// The token is the table's plain 3/3 green Elephant row. Flashback is the
// shared constructor (graveyard cast, price, and CR 702.34a's exile as
// the spell leaves the stack).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "ee243f81-f51c-4d9a-a396-f7cef84b46c1",
		Name:             "Call of the Herd",
		Completeness:     CompletenessFull,
		Purpose:          game.Purpose{Tokens: 1, CreatureTokens: 1},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{3}{G}")},
		OnResolve:        ResolveSteps(CreateToken{Template: TokenCard("3/3 green Elephant"), N: 1}),
	})
}
