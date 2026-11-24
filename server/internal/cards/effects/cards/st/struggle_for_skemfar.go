package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Struggle for Skemfar — Sorcery {3}{G} (Kaldheim):
//
//	"Put a +1/+1 counter on target creature you control. Then that creature
//	 fights up to one target creature you don't control. (Each deals damage
//	 equal to its power to the other.)
//	 Foretell {G} (During your turn, you may pay {2} and exile this card from
//	 your hand face down. Cast it on a later turn for its foretell cost.)"
//
// Two target clauses (slot 0 yours, slot 1 up to one you don't control); the
// counter comes first and stays even when the fight does not happen, so the
// fight reads the bigger power (CR 701.14). If your creature is gone, neither
// half happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "a3bdb026-e09d-443f-ba02-fe0d5412fc97",
		Name:           "Struggle for Skemfar",
		Completeness:   effects.CompletenessFull,
		SpecialActions: []game.SpecialAction{effects.Foretell("{G}")},
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("up to one target creature you don't control", effects.OpponentControls()).WithCount(0, 1),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			mine, ok := ctx.ClauseTarget(0)
			if !ok || mine.Kind != game.TargetCard {
				return nil
			}
			if err := (effects.AddCounter{Target: mine.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
				return err
			}
			theirs, ok := ctx.ClauseTarget(1)
			if !ok || theirs.Kind != game.TargetCard {
				return nil
			}
			return effects.B10Fight(ctx, mine.ID, theirs.ID)
		},
	})
}
