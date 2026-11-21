package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eldritch Immunity — Kindred Instant — Eldrazi {C} (Modern Horizons 3
// Commander, #5468):
//
//	"Target creature you control gains protection from each color until
//	 end of turn.
//	 Overload {4}{C} (You may cast this spell for its overload cost. If
//	 you do, change "target" in its text to "each.")"
//
// "Protection from each color" is five protection abilities, one per
// colour (CR 702.16), granted until end of turn. Cast normally it targets
// one creature you control; overloaded it has no target and grants the
// protection to each creature you control as it resolves (the set is
// locked in then, so a creature that enters later does not get it).
// Both costs need colorless mana, {C}, which only a source of colorless
// mana can pay.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7dce1e8c-5cfd-4626-9f23-7bb3f1b10c1e",
		Name:         "Eldritch Immunity",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerProtect, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "protection"))},
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		AlternativeCosts: []game.AlternativeCost{
			effects.Overload("{4}{C}"),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if ctx.PaidAltCost("overload") {
				return effects.GrantKeywordUntilEOT{
					Match:    effects.And(effects.Creature(), effects.YouControl()),
					Keywords: effects.OdyProtectionFromEachColor(),
					Label:    "Eldritch Immunity — protection from each color",
				}.Apply(ctx)
			}
			target := effects.FirstLegalCardTarget(ctx)
			if target == uuid.Nil {
				return nil
			}
			return effects.GrantKeywordUntilEOT{
				Target:   target,
				Keywords: effects.OdyProtectionFromEachColor(),
				Label:    "Eldritch Immunity — protection from each color",
			}.Apply(ctx)
		},
	})
}
