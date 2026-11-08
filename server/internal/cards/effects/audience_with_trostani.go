package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Audience with Trostani — Sorcery {2}{G} (Murders at Karlov Manor,
// #4469):
//
//	"Create a 0/1 green Plant creature token, then draw cards equal to
//	 the number of differently named creature tokens you control."
//
// Two steps (ResolveSteps), so the draw waits for the Plant: a token
// creation a replacement pauses is finished before the count is made,
// and the new Plant counts. Names are compared as every other name
// comparison in the catalog does (EachDifferentName): the effective
// name, case and spaces aside. A token with no name (a face-down one)
// has no name to differ, so it adds nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cbbb5c0c-e953-410d-88c6-8fedb719e94c",
		Name:         "Audience with Trostani",
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			CreateToken{Template: TokenCard("0/1 green Plant"), N: 1},
			Instruction(func(ctx *Context) error {
				return DrawCards{N: differentlyNamedCreatureTokensYouControl(ctx.Game, ctx.Controller())}.Apply(ctx)
			}),
		),
	})
}

// differentlyNamedCreatureTokensYouControl counts the distinct names
// among the creature tokens `player` controls.
func differentlyNamedCreatureTokensYouControl(g *game.Game, player uuid.UUID) int {
	names := map[string]bool{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != player || !c.IsToken() || !c.IsCreature() {
			continue
		}
		if name := strings.ToLower(strings.TrimSpace(c.Effective().Name)); name != "" {
			names[name] = true
		}
	}
	return len(names)
}
