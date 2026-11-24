package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Graceful Restoration — Sorcery {3}{W}{B} (Modern Horizons 2, #5501):
//
//	"Choose one —
//	 • Return target creature card from your graveyard to the battlefield
//	   with an additional +1/+1 counter on it.
//	 • Return up to two target creature cards with power 2 or less from your
//	   graveyard to the battlefield."
//
// The first bullet's counter is an entry counter (CR 122.6, 614.1c), so it
// is there as the creature enters and Doubling Season sees it. The second
// returns both cards together as one entry (CR 603.6a), and a card that left
// the graveyard in response is skipped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "59bfeacc-f80b-49fc-8a35-f7e1ad905161",
		Name:         "Graceful Restoration",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Return target creature card from your graveyard to the battlefield with an additional +1/+1 counter on it.",
				effects.TargetCardInGraveyard("target creature card from your graveyard", effects.Creature(), effects.YouOwn()),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					return ctx.Game.ReturnFromGraveyardWithCountersKeyedForEffect(t.ID, ctx.Controller(), false,
						map[string]int{game.CounterPlusOne: 1}, game.KeyedTail{})
				}),
			effects.ModeDoing("Return up to two target creature cards with power 2 or less from your graveyard to the battlefield.",
				effects.TargetCardInGraveyard("up to two target creature cards with power 2 or less from your graveyard",
					effects.Creature(), effects.YouOwn(), effects.PowerLE(2)).WithCount(0, 2),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					var ids []uuid.UUID
					for _, t := range ctx.ModeTargets(occ) {
						if t.Kind == game.TargetCard && ctx.IsTargetLegal(t) {
							ids = append(ids, t.ID)
						}
					}
					return effects.ReturnFromGraveyardTogether{Targets: ids, Controller: ctx.Controller()}.Apply(ctx)
				}),
		),
	})
}
