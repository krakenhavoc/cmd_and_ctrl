package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Barreling Attack — Instant {2}{R}{R} (Mirage, S82 OLD b797, #5633):
//
//	"Target creature gains trample until end of turn. When that creature
//	 becomes blocked this turn, it gets +1/+1 until end of turn for each
//	 creature blocking it."
//
// The trample is a layer-6 grant ending at cleanup. The second sentence is
// a delayed triggered ability the spell creates as it resolves (CR 603.7a),
// bound to that one object (CR 400.7): if the creature left and came back it
// is a new object and its being blocked no longer boosts it. It fires once
// (CR 603.7b) and ends with the turn; the blockers are counted as it
// resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f44eab45-8ff9-40eb-9919-28fe5203fc48",
		Name:         "Barreling Attack",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "trample")),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || !effects.OnBattlefield(ctx.Game, t.ID) {
				return nil
			}
			if err := (effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"trample"}, Label: "Barreling Attack — trample"}).Apply(ctx); err != nil {
				return err
			}
			c, ok := effects.BattlefieldCardForEffect(ctx.Game, t.ID)
			if !ok {
				return nil
			}
			return effects.DelayedOnEvent{
				Label:      "Barreling Attack — it gets +1/+1 for each creature blocking it",
				On:         []game.EventKind{game.EventBecomesBlocked},
				Condition:  effects.BarrelingAttackBlockedCondition,
				CondParams: game.EffectParams{Object: game.ObjectRef{ID: t.ID, Epoch: c.ObjectEpoch}},
				Body:       effects.BarrelingAttackBoostBody,
			}.Apply(ctx)
		},
	})
}
