package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Elephant Ambush — Instant {2}{G}{G} (Odyssey, #5613):
//
//	"Create a 3/3 green Elephant creature token.
//	 Flashback {6}{G}{G} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// Call of the Herd's token at instant speed: the table's plain 3/3 green
// Elephant row (CR 701.7a).
//
// Flashback is the shared constructor: the graveyard cast, the price,
// and CR 702.34a's exile as the spell leaves the stack (resolved,
// countered, or with its target gone).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "09386931-59bd-481f-bccc-32e8a8abc0f9",
		Name:             "Elephant Ambush",
		Completeness:     CompletenessFull,
		Purpose:          game.Purpose{Tokens: 1, CreatureTokens: 1},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{6}{G}{G}")},
		OnResolve:        ResolveSteps(CreateToken{Template: TokenCard("3/3 green Elephant"), N: 1}),
	})
}
