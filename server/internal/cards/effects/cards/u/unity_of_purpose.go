package cardsu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unity of Purpose — Instant {3}{U} (Oath of the Gatewatch, #5545):
//
//	"Support 2. (Put a +1/+1 counter on each of up to two target creatures.)
//	 Untap each creature you control with a +1/+1 counter on it."
//
// Support from an instant may target any creatures (CR 701.41a; Lead by
// Example's clause). The untap is a second sentence read as the spell resolves,
// after the counters, so a creature that just got one untaps: every creature
// its controller controls with a +1/+1 counter on it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9fbcae6e-5523-47a3-b0ff-ab7a8bb72822",
		Name:         "Unity of Purpose",
		Completeness: effects.CompletenessFull,
		Targets:      effects.SupportTargets(2, false),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.SupportEffect(ctx.Game, item); err != nil {
				return err
			}
			for _, c := range effects.MatchingBattlefield(ctx, effects.And(effects.Creature(), effects.YouControl(),
				func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.Counters[game.CounterPlusOne] > 0 })) {
				if err := (effects.UntapTarget{Target: c.InstanceID}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
