package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Seismic Shift — Sorcery {3}{R} (Dominaria, S82 OLD batch 214, #5530):
//
//	"Destroy target land. Up to two target creatures can't block this turn."
//
// Two target clauses (CR 115.1): slot 0 the land, slot 1 up to two creatures.
// Each part works on the targets still legal as the spell resolves
// (CR 608.2b): an illegal land is not destroyed but the creatures still can't
// block, and the other way round; only a spell with no legal target at all is
// countered. The restriction is Beat a Path's until-end-of-turn "can't block"
// record (CR 509.1b).
//
// No simplification.
func init() {
	const label = "Seismic Shift — can't block this turn"
	effects.Register(effects.Spec{
		OracleID:     "b81894bc-d5ca-44ee-b837-bdb8fec32881",
		Name:         "Seismic Shift",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Targets: game.ForTargets(effects.CantBlockUntilEOT(1, 0, 0))},
		Targets: effects.Clauses(
			effects.TargetPermanent("target land", effects.Land()),
			effects.TargetCreature("up to two target creatures").WithCount(0, 2),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if land, ok := ctx.ClauseTarget(0); ok && land.Kind == game.TargetCard && ctx.IsTargetLegal(land) {
				if err := (effects.DestroyTarget{Target: land.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			for _, t := range ctx.ClauseTargets(1) {
				if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
					continue
				}
				if err := (effects.RestrictUntilEOT{Target: t.ID, Restrictions: game.CantBlock, Label: label}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
