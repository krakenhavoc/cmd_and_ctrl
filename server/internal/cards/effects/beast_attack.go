package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Beast Attack — Instant {2}{G}{G}{G} (Odyssey, #5613):
//
//	"Create a 4/4 green Beast creature token.
//	 Flashback {2}{G}{G}{G} (You may cast this card from your graveyard
//	 for its flashback cost. Then exile it.)"
//
// The token is the table's plain 4/4 green Beast row. Flashback is the
// shared constructor (graveyard cast, price, and CR 702.34a's exile as
// the spell leaves the stack); the flashed-back spell is still an
// instant, so it can be cast at instant speed from the graveyard too.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "3f5acd90-199c-4ed3-9c8a-eeac4670d876",
		Name:             "Beast Attack",
		Completeness:     CompletenessFull,
		Purpose:          game.Purpose{Tokens: 1, CreatureTokens: 1},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{2}{G}{G}{G}")},
		OnResolve:        ResolveSteps(CreateToken{Template: TokenCard("4/4 green Beast"), N: 1}),
	})
}
