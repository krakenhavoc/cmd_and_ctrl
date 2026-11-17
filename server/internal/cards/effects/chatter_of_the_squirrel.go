package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chatter of the Squirrel — Sorcery {G} (Odyssey, #5613):
//
//	"Create a 1/1 green Squirrel creature token.
//	 Flashback {1}{G} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// The token is the table's plain 1/1 green Squirrel row. Flashback is the
// shared constructor (graveyard cast, price, and CR 702.34a's exile as
// the spell leaves the stack).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "b0aaa4b1-5188-43a6-997d-7a9b2ad452bc",
		Name:             "Chatter of the Squirrel",
		Completeness:     CompletenessFull,
		Purpose:          game.Purpose{Tokens: 1, CreatureTokens: 1},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{1}{G}")},
		OnResolve:        ResolveSteps(CreateToken{Template: TokenCard("1/1 green Squirrel"), N: 1}),
	})
}
