package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vertigo — Instant {R} (Ice Age, S82 OLD batch 840, #5636):
//
//	"Vertigo deals 2 damage to target creature with flying. That creature loses
//	 flying until end of turn."
//
// The damage, then a layer 6 removal of flying with its own timestamp that ends at
// cleanup (CR 514.2), as Whiteout's is. A creature the damage destroyed is gone and
// loses nothing; a creature that survives it can't block fliers, nor fly past them.
// "With flying" is judged as the spell is cast and again as it resolves (CR 608.2b),
// through the layered view.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1dabfade-79e5-44e2-afa6-b20498c6bb8b",
		Name:         "Vertigo",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature with flying", effects.HasKeyword("flying")),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 2)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.DealDamage{Source: ctx.Source(), Target: id, Amount: 2}).Apply(ctx); err != nil {
				return err
			}
			if !effects.OnBattlefield(ctx.Game, id) {
				return nil
			}
			return effects.UntilEndOfTurn(ctx, id, nil, "Vertigo — loses flying", game.RemoveKeywordsMod("flying"))
		},
	})
}
