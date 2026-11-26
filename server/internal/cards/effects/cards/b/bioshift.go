package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bioshift — Instant {G/U} (Gatecrash, S82 OLD #5565):
//
//	"Move any number of +1/+1 counters from target creature onto another
//	 target creature with the same controller."
//
// Simic Guildmage's two clauses (the second is related to the first by
// controller, ControlledByTargetsController, and distinct from it) with
// Scrounging Bandar's amount: the number to move is chosen as the spell
// resolves, from none up to the counters the first creature holds then
// (CR 608.2c). Moving counters is removing them and putting them on the other
// creature, and if either half can't happen neither does (CR 122.5), so both
// creatures must still be legal targets and on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3da604fc-68e9-4749-98f9-6fbcebcab9b6",
		Name:         "Bioshift",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("target creature"),
			effects.Distinct(effects.ControlledByTargetsController(effects.TargetCreature("another target creature with the same controller"), 0)),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			from, okFrom := ctx.ClauseTarget(0)
			to, okTo := ctx.ClauseTarget(1)
			if !okFrom || !okTo || from.Kind != game.TargetCard || to.Kind != game.TargetCard {
				return nil
			}
			src, ok := ctx.Game.LookupCardForEffect(from.ID)
			if !ok || src.Counters[game.CounterPlusOne] <= 0 {
				return nil
			}
			return effects.ChooseNumber{
				Min:      0,
				Max:      src.Counters[game.CounterPlusOne],
				Question: "Bioshift — how many +1/+1 counters to move onto the other creature?",
				Unit:     game.PayAmountCounters,
				Then:     bioshiftMoved,
				Carry:    []uuid.UUID{from.ID, to.ID},
			}.Apply(ctx)
		},
	})
}

// bioshiftMoved moves the chosen number, only while both creatures are still on
// the battlefield (CR 122.5). An on-disk key: never renamed, never reused.
var bioshiftMoved = effects.NumberThen("number/s82-old-b411-bioshift-move",
	func(ctx *effects.Context, r game.NumberAnswer) error {
		if len(r.Carry) < 2 || r.Number <= 0 {
			return nil
		}
		from, to := r.Carry[0], r.Carry[1]
		if !effects.OnBattlefield(ctx.Game, from) || !effects.OnBattlefield(ctx.Game, to) {
			return nil
		}
		moved, err := effects.SocB9RemoveCounters(ctx, from, game.CounterPlusOne, r.Number)
		if err != nil || moved <= 0 {
			return err
		}
		return effects.AddCounter{Target: to, Kind: game.CounterPlusOne, N: moved}.Apply(ctx)
	})
