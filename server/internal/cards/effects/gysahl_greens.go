package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gysahl Greens — Sorcery {1}{G} (Final Fantasy, #4462):
//
//	"Create a 2/2 green Bird creature token with "Whenever a land you
//	 control enters, this token gets +1/+0 until end of turn."
//	 Flashback {6}{G} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// Bestial Incursion's shape with the set's Chocobo Bird
// (chocoboBirdToken), whose landfall is printed on the token. Flashback
// is the shared constructor: the graveyard cast zone, the price and the
// exile as it leaves the stack (CR 702.34a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "eddfe7d7-6868-4b81-b46c-d0765e5210e7",
		Name:             "Gysahl Greens",
		Completeness:     CompletenessFull,
		Purpose:          game.Purpose{Tokens: 1},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{6}{G}")},
		OnResolve:        ResolveSteps(CreateToken{Template: chocoboBirdToken(), N: 1}),
	})
}
