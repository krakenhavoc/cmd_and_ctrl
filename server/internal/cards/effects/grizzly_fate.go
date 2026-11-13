package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grizzly Fate — Sorcery {3}{G}{G} (Judgment, #5611):
//
//	"Create two 2/2 green Bear creature tokens.
//	 Threshold — Create four 2/2 green Bear creature tokens instead if
//	 there are seven or more cards in your graveyard.
//	 Flashback {5}{G}{G} (You may cast this card from your graveyard for
//	 its flashback cost. Then exile it.)"
//
// The existing 2/2 green Bear row, made together. Threshold is an
// ability word (CR 207.2c) read as the spell resolves: the spell itself
// is on the stack then, not in the graveyard, so it never counts toward
// its own seven. Flashback is the shared constructor, exiled as it
// leaves the stack (CR 702.34a). The Purpose declares the two Bears the
// spell always makes.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "f3073e4f-ca7e-4fde-88a4-bd7b6ebfbc8c",
		Name:             "Grizzly Fate",
		Completeness:     CompletenessFull,
		Purpose:          game.Purpose{Tokens: 2, CreatureTokens: 2},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{5}{G}{G}")},
		OnResolve: ResolveSteps(Instruction(func(ctx *Context) error {
			n := 2
			if fraThresholdMet(ctx.Game, ctx.Controller()) {
				n = 4
			}
			return CreateToken{Template: TokenCard("2/2 green Bear"), N: n}.Apply(ctx)
		})),
	})
}
