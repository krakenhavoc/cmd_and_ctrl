package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crush of Wurms — Sorcery {6}{G}{G}{G} (Judgment, #5611, #5789):
//
//	"Create three 6/6 green Wurm creature tokens.
//	 Flashback {9}{G}{G}{G} (You may cast this card from your graveyard
//	 for its flashback cost. Then exile it.)"
//
// Roar of the Wurm at three tokens: the table's plain 6/6 green Wurm
// row, and the shared Flashback constructor (graveyard cast, price, and
// CR 702.34a's exile as the spell leaves the stack).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "6c12cd05-b673-453c-9f6b-f3aa022467c1",
		Name:             "Crush of Wurms",
		Completeness:     CompletenessFull,
		Purpose:          game.Purpose{Tokens: 3, CreatureTokens: 3},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{9}{G}{G}{G}")},
		OnResolve:        ResolveSteps(CreateToken{Template: TokenCard("6/6 green Wurm"), N: 3}),
	})
}
