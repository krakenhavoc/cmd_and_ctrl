package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Old Man of the Sea — Creature — Djinn {1}{U}{U}, 2/3:
//
//	"You may choose not to untap this creature during your untap step.
//	 {T}: Gain control of target creature with power less than or equal to this creature's power for as long as this creature remains tapped and that creature's power remains less than or equal to this creature's power."
//
// #1863 and ADR 0109 §3: the target is "with power less than or equal
// to this creature's power", a clause relative to the source
// (PowerNoGreater, #2146), judged as the target is chosen and again as
// the ability resolves. The duration is the conjunction ADR 0109 added
// for it (Duration.Also): the control ends the moment the Old Man
// untaps or leaves, or the creature's power rises above his,
// re-checked after every layer pass. The Old Man himself qualifies as a
// target, as printed; stealing him changes nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cce84cf1-5574-43b0-9d75-72e6451403a7",
		Name:         "Old Man of the Sea",
		Completeness: effects.CompletenessFull,
		UntapOptOuts: []game.UntapOptOut{effects.MayChooseNotToUntapSelf("Old Man of the Sea — you may choose not to untap this creature")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Gain control of target creature with power less than or equal to this creature's power for as long as this creature remains tapped and that creature's power remains less than or equal to this creature's power.",
			Cost:    effects.TapCost(),
			Targets: effects.RelativeToSource(effects.TargetCreature("target creature with power less than or equal to this creature's power"), effects.PowerNoGreater()),
			Effect:  oldManOfTheSeaSteals,
		}},
	})
}

// oldManOfTheSeaSteals gains control of the legal target for as long as
// the source stays tapped and the target's power stays at most the
// source's. A duration already false as the ability resolves (the Old
// Man untapped, left, or the creature grew) takes nothing (CR 611.2b).
func oldManOfTheSeaSteals(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	target := effects.FirstLegalBattlefieldTarget(ctx)
	if target == uuid.Nil || ctx.IsNewSourceObjectAsThis(ctx.Source()) || effects.SourceUntappedSinceActivation(ctx) { // #1432, #4061
		return nil
	}
	d, ok := g.ForAsLongAsSourceTappedAndPowerAtMostDuration(ctx.Source(), target)
	if !ok {
		return nil
	}
	return effects.GainControl{
		Target: target, Controller: ctx.Controller(), Duration: d,
		Label: "Old Man of the Sea — control while he is tapped and its power stays at most his",
	}.Apply(ctx)
}
