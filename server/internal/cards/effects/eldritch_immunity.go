package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "7dce1e8c-5cfd-4626-9f23-7bb3f1b10c1e",
		Name:         "Eldritch Immunity",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerProtect, Targets: game.ForTargets(BenefitUntilEOT(0, 0, 0, "protection"))},
		Targets:      TargetCreature("target creature you control", YouControl()),
		AlternativeCosts: []game.AlternativeCost{
			Overload("{4}{C}"),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if ctx.PaidAltCost("overload") {
				return GrantKeywordUntilEOT{
					Match:    And(Creature(), YouControl()),
					Keywords: odyProtectionFromEachColor(),
					Label:    "Eldritch Immunity — protection from each color",
				}.Apply(ctx)
			}
			target := firstLegalCardTarget(ctx)
			if target == uuid.Nil {
				return nil
			}
			return GrantKeywordUntilEOT{
				Target:   target,
				Keywords: odyProtectionFromEachColor(),
				Label:    "Eldritch Immunity — protection from each color",
			}.Apply(ctx)
		},
	})
}
