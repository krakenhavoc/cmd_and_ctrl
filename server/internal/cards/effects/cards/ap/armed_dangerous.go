package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Armed // Dangerous — split card, oracle a0593a5b-6f76-4eb3-a1b1-9794d2d36c92
// (Dragon's Maze, S82 OLD batch 402, #5564):
//
//	Armed     — Sorcery {1}{R}: "Target creature gets +1/+1 and gains double
//	            strike until end of turn."
//	Dangerous — Sorcery {3}{G}: "All creatures able to block target creature this
//	            turn do so."
//	            Fuse (You may cast one or both halves of this card from your hand.)
//
// Both halves are ADR 0034 faces (Armed under the bare oracle ID, Dangerous under
// "#1"); the fused cast is the engine's (CR 702.102). Armed is a layer-7 pump and
// a layer-6 keyword grant, both until cleanup. Dangerous is Bloodscent's lure
// (Alluring Scent's block requirement, CR 509.1c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a0593a5b-6f76-4eb3-a1b1-9794d2d36c92",
		Name:         "Armed",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1, "double strike")),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: "Armed — +1/+1"}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"double strike"}, Label: "Armed — double strike"}.Apply(ctx)
		},
	})
	effects.Register(effects.Spec{
		OracleID:     "a0593a5b-6f76-4eb3-a1b1-9794d2d36c92#1",
		Name:         "Dangerous",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.LureTargetForTheTurn(ctx, "Dangerous — all creatures able to block it do so")
		},
	})
}
