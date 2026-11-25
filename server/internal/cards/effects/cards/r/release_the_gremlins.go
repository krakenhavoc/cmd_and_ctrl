package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Release the Gremlins — {X}{X}{R} Sorcery (Aether Revolt, #5536):
//
//	"Destroy X target artifacts. Create X 2/2 red Gremlin creature tokens."
//
// X is announced (CR 107.3b) and sets both the number of targets and the
// number of tokens. A target that has become illegal is skipped; the
// Gremlins are made for the full X as long as the spell resolves at all
// (CR 608.2b).
//
// No simplification.
func init() {
	spec := effects.TargetPermanent("X target artifacts", effects.Artifact())
	spec.CountFromX = true
	effects.Register(effects.Spec{
		OracleID:     "3b583358-d0d5-4f53-b06c-c6439156e684",
		Name:         "Release the Gremlins",
		XMatters:     true,
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Tokens: 1, CreatureTokens: 1},
		Targets:      spec,
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.DestroyEachLegalTarget(ctx.Item, ctx) }),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.CreateToken{Template: effects.TokenCard("2/2 red Gremlin"), N: ctx.X()}.Apply(ctx)
			}),
		),
	})
}
