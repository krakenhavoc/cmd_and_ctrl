package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Weapon Surge — Instant {R} (Dragon's Maze, S82 OLD batch 404, #5564):
//
//	"Target creature you control gets +1/+0 and gains first strike until end of
//	 turn.
//	 Overload {1}{R} (You may cast this spell for its overload cost. If you do,
//	 change "target" in its text to "each.")"
//
// Scale Up's shape: one effect until end of turn, a layer-7 +1/+0 and a layer-6
// first strike grant, on the target creature you control or, overloaded (CR
// 702.96, the engine's: it clears the target), on each creature you control, the
// set fixed as the spell resolves (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2c46bc05-1133-4eaf-8865-eb8a1d880bd9",
		Name:         "Weapon Surge",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0, "first strike")),
		AlternativeCosts: []game.AlternativeCost{
			effects.CostWithPurpose(effects.Overload("{1}{R}"), game.Purpose{Answers: game.AnswerCombatGrant}),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			mods := []game.Mod{game.ModifyPTMod(1, 0), game.AddKeywordsMod("first strike")}
			const label = "Weapon Surge — +1/+0 and first strike until end of turn"
			if ctx.PaidAltCost("overload") {
				return effects.UntilEndOfTurn(ctx, uuid.Nil, effects.And(effects.Creature(), effects.YouControl()), label, mods...)
			}
			id := effects.FirstLegalBattlefieldTarget(ctx)
			if id == uuid.Nil {
				return nil
			}
			return effects.UntilEndOfTurn(ctx, id, nil, label, mods...)
		},
	})
}
